package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/noders-team/go-daml/pkg/client"
	"github.com/noders-team/go-daml/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	grpcAddress := os.Getenv("GRPC_ADDRESS")
	if grpcAddress == "" {
		grpcAddress = "localhost:8080"
	}

	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		log.Warn().Msg("BEARER_TOKEN environment variable not set")
	}

	tlsConfig := client.TlsConfig{}

	cl, err := client.NewDamlClient(bearerToken, grpcAddress).
		WithTLSConfig(tlsConfig).
		Build(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build DAML client")
	}

	darFilePath := "./test-data/all-kinds-of-1.0.0.dar"
	darContent, err := os.ReadFile(darFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read DAR file")
	}

	uploadedPackageName := "all-kinds-of"
	if !packageExists(uploadedPackageName, cl) {
		log.Info().Msg("skipping package upload as it already exists")

		submissionID := "validate-" + time.Now().Format("20060102150405")
		log.Info().Str("submissionID", submissionID).Msg("validating DAR file")

		err = cl.PackageMng.ValidateDarFile(context.Background(), darContent, submissionID)
		if err != nil {
			log.Fatal().Err(err).Msgf("DAR validation failed for %s", darFilePath)
		}

		uploadSubmissionID := "upload-" + time.Now().Format("20060102150405")
		log.Info().Str("submissionID", uploadSubmissionID).Msg("uploading DAR file")

		err = cl.PackageMng.UploadDarFile(context.Background(), darContent, uploadSubmissionID)
		if err != nil {
			log.Fatal().Err(err).Msg("DAR upload failed")
		}

		if !packageExists(uploadedPackageName, cl) {
			log.Fatal().Msg("package not found")
		}
	}

	packageID := "ddf0d6396a862eaa7f8d647e39d090a6b04c4a3fd6736aa1730ebc9fca6be664"
	status, err := cl.PackageService.GetPackageStatus(context.Background(),
		&model.GetPackageStatusRequest{PackageID: packageID})
	if err != nil {
		log.Fatal().Err(err).Str("packageId", packageID).Msg("failed to get package status")
	}
	log.Info().Msgf("package status: %v", status.PackageStatus)

	parties, err := cl.PartyMng.ListKnownParties(context.Background(), "", 0, "")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list parties")
	}

	for _, party := range parties.PartyDetails {
		log.Info().Msgf("party: %+v", party)
	}

	participantID, err := cl.PartyMng.GetParticipantID(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list parties")
	}
	log.Info().Msgf("participantID: %s", participantID)

	// party := "test_party_artem_bogomaz_::1220716cdae4d7884d468f02b30eb826a7ef54e98f3eb5f875b52a0ef8728ed98c3a"

	mappyContract := MappyContract{
		Operator: "Alice",
		Value: GENMAP{
			"a": "b",
			"c": "d",
		},
	}

	// Create Archive command
	archiveCmd := mappyContract.Archive("contract-id-123")

	// Submit the Archive command
	commandID := "archive-" + time.Now().Format("20060102150405")
	submissionReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			WorkflowID:   "archive-workflow-" + time.Now().Format("20060102150405"),
			CommandID:    commandID,
			ActAs:        []string{participantID},
			SubmissionID: "sub-" + time.Now().Format("20060102150405"),
			DeduplicationPeriod: model.DeduplicationDuration{
				Duration: 60 * time.Second,
			},
			Commands: []*model.Command{{Command: archiveCmd}},
		},
	}

	response, err := cl.CommandService.SubmitAndWait(context.Background(), submissionReq)
	if err != nil {
		log.Fatal().Err(err).Str("packageId", packageID).Msg("failed to get package status")
	}
	log.Info().Msgf("response.UpdateID: %s", response.UpdateID)
}

func packageExists(pkgName string, cl *client.DamlBindingClient) bool {
	updatedPackages, err := cl.PackageMng.ListKnownPackages(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("failed to list packages after upload")
		return false
	}

	for _, pkg := range updatedPackages {
		if strings.EqualFold(pkg.Name, pkgName) {
			log.Info().Msgf("package already exists %+v", pkg)
			return true
		}
	}

	return false
}
