package main

import (
	"context"
	"fmt"
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

	log.Info().Str("generatedPackageID", PackageID).Msg("Using package ID from generated code")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

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
		log.Info().Msg("package not found, uploading")

		submissionID := "validate-" + time.Now().Format("20060102150405")
		log.Info().Str("submissionID", submissionID).Msg("validating DAR file")

		err = cl.PackageMng.ValidateDarFile(ctx, darContent, submissionID)
		if err != nil {
			log.Fatal().Err(err).Msgf("DAR validation failed for %s", darFilePath)
		}

		uploadSubmissionID := "upload-" + time.Now().Format("20060102150405")
		log.Info().Str("submissionID", uploadSubmissionID).Msg("uploading DAR file")

		err = cl.PackageMng.UploadDarFile(ctx, darContent, uploadSubmissionID)
		if err != nil {
			log.Fatal().Err(err).Msg("DAR upload failed")
		}

		if !packageExists(uploadedPackageName, cl) {
			log.Fatal().Msg("package not found")
		}
	}
	status, err := cl.PackageService.GetPackageStatus(ctx,
		&model.GetPackageStatusRequest{PackageID: PackageID})
	if err != nil {
		log.Fatal().Err(err).Str("packageId", PackageID).Msg("failed to get package status")
	}
	log.Info().Msgf("package status: %v", status.PackageStatus)

	party := "app_provider_localnet-localparty-1::1220716cdae4d7884d468f02b30eb826a7ef54e98f3eb5f875b52a0ef8728ed98c3a"

	contractIDs, err := createContract(ctx, party, cl)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create contract")
	}

	// Create MappyContract
	mappyContract := MappyContract{
		Operator: PARTY(party),
		Value: GENMAP{
			"key1": "value1",
			"key2": "value2",
		},
	}

	/*
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		oneOfEverythingTemplateID := fmt.Sprintf("%s:%s:%s", PackageID, "AllKindsOf", "OneOfEverything")
		mappyContractTemplateID := fmt.Sprintf("%s:%s:%s", PackageID, "AllKindsOf", "MappyContract")

		log.Info().Msg("searching for ALL active contracts for the party...")
		contractsCh, errCh := cl.StateService.GetActiveContracts(ctx,
			&model.GetActiveContractsRequest{
				Filter: &model.TransactionFilter{
					FiltersByParty: map[string]*model.Filters{
						party: {
							Inclusive: &model.InclusiveFilters{
								TemplateFilters: []*model.TemplateFilter{
									{
										TemplateID:              oneOfEverythingTemplateID,
										IncludeCreatedEventBlob: true,
									},
									{
										TemplateID:              mappyContractTemplateID,
										IncludeCreatedEventBlob: true,
									},
								},
							},
						},
					},
				},
				Verbose: true,
			})

		contractCount := 0
		var firstContractID string
		done := false
		for !done {
			select {
			case response, ok := <-contractsCh:
				if !ok {
					log.Info().Int("totalContracts", contractCount).Msg("active contracts stream completed")
					cancel()
					done = true
					break
				}
				if response != nil && len(response.ActiveContracts) > 0 {
					contractCount += len(response.ActiveContracts)
					log.Info().
						Int("activeContracts", len(response.ActiveContracts)).
						Int64("offset", response.Offset).
						Msg("received active contracts batch")

					// Log details of each contract found
					for i, contract := range response.ActiveContracts {
						log.Info().
							Int("contractIndex", i).
							Str("contractID", contract.ContractID).
							Str("templateID", contract.TemplateID).
							Msg("Found active contract")

						if firstContractID == "" {
							firstContractID = contract.ContractID
							log.Info().Str("contractID", firstContractID).Msg("captured first contract ID for archive example")
						}
					}
				} else if response != nil {
					log.Info().
						Int64("offset", response.Offset).
						Msg("received empty contracts batch")
				}
			case err := <-errCh:
				if err != nil {
					log.Error().Err(err).Msg("active contracts stream error")
					cancel()
					done = true
					break
				}
			case <-ctx.Done():
				log.Info().Int("totalContracts", contractCount).Msg("active contracts stream timeout")
				done = true
				break
			}
		}

		log.Info().Msg("finished reading active contracts")
		if contractCount == 0 {
			log.Warn().Msgf("no active contracts found for party %s with our specific template filters", party)

			// Try searching for ANY active contracts for this party
			log.Info().Msg("Searching for ANY active contracts for the party...")
			ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel2()

			allContractsCh, allErrCh := cl.StateService.GetActiveContracts(ctx2,
				&model.GetActiveContractsRequest{
					Filter: &model.TransactionFilter{
						FiltersByParty: map[string]*model.Filters{
							party: {
								Inclusive: &model.InclusiveFilters{},
							},
						},
					},
					Verbose: true,
				})

			allContractCount := 0
			allDone := false
			for !allDone {
				select {
				case response, ok := <-allContractsCh:
					if !ok {
						log.Info().Int("totalContracts", allContractCount).Msg("all contracts search completed")
						cancel2()
						allDone = true
						break
					}
					if response != nil && len(response.ActiveContracts) > 0 {
						allContractCount += len(response.ActiveContracts)
						for i, contract := range response.ActiveContracts {
							log.Info().
								Int("contractIndex", i).
								Str("contractID", contract.ContractID).
								Str("templateID", contract.TemplateID).
								Msg("Found ANY active contract")
						}
					}
				case err := <-allErrCh:
					if err != nil {
						log.Error().Err(err).Msg("all contracts search error")
						cancel2()
						allDone = true
						break
					}
				case <-ctx2.Done():
					log.Info().Msg("all contracts search timeout")
					allDone = true
					break
				}
			}

			if allContractCount == 0 {
				log.Error().Msgf("no active contracts found at all for party %s", party)
				return
			} else {
				log.Info().Msgf("found %d contracts total, but none match our templates", allContractCount)
				return
			}
		}*/

	participantID, err := cl.PartyMng.GetParticipantID(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get participant ID")
	}
	log.Info().Msgf("participantID: %s", participantID)

	users, err := cl.UserMng.ListUsers(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list users")
	}
	for _, u := range users {
		log.Info().Msgf("user: %+v", u)
	}

	rights, err := cl.UserMng.ListUserRights(ctx, "app-provider")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to list user rights")
	}
	rightsGranded := false
	for _, r := range rights {
		canAct, ok := r.Type.(model.RightType).(model.CanActAs)
		if ok && canAct.Party == party {
			rightsGranded = true
		}
	}

	if !rightsGranded {
		log.Info().Msg("grant rights")
		newRights := make([]*model.Right, 0)
		newRights = append(newRights, &model.Right{Type: model.CanReadAs{Party: party}})
		_, err = cl.UserMng.GrantUserRights(context.Background(), "app-provider", "", newRights)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to grant user rights")
		}
	}

	// mappyContract already created above for contract creation

	// Create Archive command using contract IDs from creation
	if len(contractIDs) == 0 {
		log.Warn().Msg("No contracts were created, cannot demonstrate Archive command")
		return
	}

	var firstContractID string
	if len(contractIDs) > 0 {
		firstContractID = contractIDs[0]
	}
	log.Info().Str("contractID", firstContractID).Msg("Using contract ID from creation for Archive command")
	archiveCmd := mappyContract.Archive(firstContractID)

	// Submit the Archive command
	commandID := "archive-" + time.Now().Format("20060102150405")
	submissionReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			WorkflowID:   "archive-workflow-" + time.Now().Format("20060102150405"),
			CommandID:    commandID,
			ActAs:        []string{party},
			SubmissionID: "sub-" + time.Now().Format("20060102150405"),
			DeduplicationPeriod: model.DeduplicationDuration{
				Duration: 60 * time.Second,
			},
			Commands: []*model.Command{{Command: archiveCmd}},
		},
	}

	response, err := cl.CommandService.SubmitAndWait(ctx, submissionReq)
	if err != nil {
		log.Fatal().Err(err).Str("packageId", PackageID).Msg("failed to submit and wait")
	}
	log.Info().Msgf("response.UpdateID: %s", response.UpdateID)

	time.Sleep(5 * time.Second)
	respUpd, err := cl.UpdateService.GetTransactionByID(ctx, &model.GetTransactionByIDRequest{
		UpdateID:          response.UpdateID,
		RequestingParties: []string{party},
	})
	if err != nil {
		log.Fatal().Err(err).Str("packageId", PackageID).Msg("failed to GetTransactionByID")
	}
	if respUpd.Transaction != nil {
		for _, event := range respUpd.Transaction.Events {
			if exercisedEvent := event.Exercised; exercisedEvent != nil {
				contractIDs = append(contractIDs, exercisedEvent.ContractID)
				log.Info().
					Str("contractID", exercisedEvent.ContractID).
					Str("templateID", exercisedEvent.TemplateID).
					Msg("found created contract in transaction")
			}
		}
	}
}

func packageExists(pkgName string, cl *client.DamlBindingClient) bool {
	updatedPackages, err := cl.PackageMng.ListKnownPackages(context.Background())
	if err != nil {
		log.Warn().Err(err).Msg("failed to list packages after upload")
		return false
	}

	for _, pkg := range updatedPackages {
		if strings.EqualFold(pkg.Name, pkgName) {
			log.Warn().Msgf("package already exists %+v", pkg)
			pkgInspect, err := cl.PackageService.GetPackage(context.Background(),
				&model.GetPackageRequest{PackageID: pkg.PackageID})
			if err != nil {
				log.Warn().Err(err).Msgf("failed to get package details for %s", pkg.Name)
				return true
			}
			log.Warn().Msgf("package details: Hash: %s HashFunction: %d", pkgInspect.Hash, pkgInspect.HashFunction)
			return true
		}
	}

	return false
}

func createContract(ctx context.Context, party string, cl *client.DamlBindingClient) ([]string, error) {
	log.Info().Msg("Creating sample contracts...")

	// Create OneOfEverything contract (commented out for now to focus on MappyContract)
	/*
		now := time.Now()
		oneOfEverythingContract := OneOfEverything{
			Operator:        PARTY(party),
			SomeBoolean:     true,
			SomeInteger:     42,
			SomeDecimal:     nil, // NUMERIC can be nil
			SomeMaybe:       nil, // OPTIONAL can be nil
			SomeMaybeNot:    nil,
			SomeText:        "Hello World",
			SomeDate:        DATE(now),
			SomeDatetime:    TIMESTAMP(now),
			SomeSimpleList:  LIST{"item1", "item2"},
			SomeSimplePair:  MyPair{Left: "left", Right: "right"},
			SomeNestedPair:  MyPair{Left: "nested-left", Right: "nested-right"},
			SomeUglyNesting: VPair{Left: func() *interface{} { var v interface{} = "ugly"; return &v }()},
			SomeMeasurement: nil,
			SomeEnum:        ColorRed,
			TheUnit:         UNIT{},
		}
	*/

	// Submit contract creation commands - start with just MappyContract to debug the issue
	createCommands := []*model.Command{
		{
			Command: &model.CreateCommand{
				TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "AllKindsOf", "MappyContract"),
				Arguments: map[string]interface{}{
					"operator": map[string]interface{}{
						"_type": "party",
						"value": party,
					},
					"value": map[string]interface{}{
						"_type": "genmap",
						"value": map[string]interface{}{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
	}

	createSubmissionReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			WorkflowID:   "create-contracts-" + time.Now().Format("20060102150405"),
			CommandID:    "create-" + time.Now().Format("20060102150405"),
			ActAs:        []string{party},
			SubmissionID: "create-sub-" + time.Now().Format("20060102150405"),
			DeduplicationPeriod: model.DeduplicationDuration{
				Duration: 60 * time.Second,
			},
			Commands: createCommands,
		},
	}

	log.Info().Msg("submitting contract creation commands...")
	createResponse, err := cl.CommandService.SubmitAndWait(context.Background(), createSubmissionReq)
	if err != nil {
		log.Err(err).Msg("failed to create contracts")
		return nil, err
	}
	log.Info().Str("updateID", createResponse.UpdateID).Msg("Successfully created contracts")

	// Use the updateID to get transaction details and extract contract IDs
	contractIDs, err := getContractIDsFromUpdate(ctx, party, createResponse.UpdateID, cl)
	if err != nil {
		log.Err(err).Msg("failed to get contract IDs from update")
		return nil, err
	}

	log.Info().Strs("contractIDs", contractIDs).Msg("extracted contract IDs from transaction")

	return contractIDs, nil
}

func getContractIDsFromUpdate(ctx context.Context, party, updateID string, cl *client.DamlBindingClient) ([]string, error) {
	response, err := cl.UpdateService.GetTransactionByID(ctx, &model.GetTransactionByIDRequest{
		UpdateID:          updateID,
		RequestingParties: []string{party},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	var contractIDs []string
	if response.Transaction != nil {
		for _, event := range response.Transaction.Events {
			if createdEvent := event.Created; createdEvent != nil {
				contractIDs = append(contractIDs, createdEvent.ContractID)
				log.Info().
					Str("contractID", createdEvent.ContractID).
					Str("templateID", createdEvent.TemplateID).
					Msg("found created contract in transaction")
			}
		}
	}

	return contractIDs, nil
}
