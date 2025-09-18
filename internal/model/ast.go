package model

import "time"

type TmplStruct struct {
	Name   string
	Fields []*TmplField
}

type TmplField struct {
	Type string
	Name string
}

type Package struct {
	Name    string
	Version string
	// Modules  map[string]*Module
	// Metadata map[string]string
	Structs map[string]*TmplStruct
}

type Module struct {
	Name       string
	DataTypes  map[string]*DataType
	Values     map[string]*Value
	Templates  map[string]*Template
	Interfaces map[string]*Interface
}

type DataType struct {
	Name         string
	TypeParams   []string
	DataCons     *DataCons
	Serializable bool
	Location     *Location
}

type DataCons struct {
	Record  *Record
	Variant *Variant
	Enum    *Enum
}

type Record struct {
	Fields []*Field
}

type Variant struct {
	Constructors []*Constructor
}

type Enum struct {
	Constructors []string
}

type Field struct {
	Name string
	Type *Type
}

type Constructor struct {
	Name string
	Type *Type
}

type Type struct {
	Var    *TypeVar
	Con    *TypeCon
	Prim   *PrimType
	Fun    *FunType
	App    *AppType
	Forall *ForallType
	Struct *StructType
	Nat    *int64
	Syn    *SynType
}

type TypeVar struct {
	Var string
}

type TypeCon struct {
	Tycon string
}

type PrimType struct {
	Prim string
}

type FunType struct {
	Params []*Type
	Result *Type
}

type AppType struct {
	Tyfun *Type
	Args  []*Type
}

type ForallType struct {
	Vars []*TypeVarWithKind
	Body *Type
}

type StructType struct {
	Fields []*FieldWithType
}

type SynType struct {
	Tysyn *TypeSynName
	Args  []*Type
}

type TypeVarWithKind struct {
	Var  string
	Kind *Kind
}

type FieldWithType struct {
	Field string
	Type  *Type
}

type TypeSynName struct {
	Name string
}

type Kind struct {
	Star  bool
	Arrow *ArrowKind
	Nat   bool
}

type ArrowKind struct {
	Params []*Kind
	Result *Kind
}

type Template struct {
	Name         string
	Param        string
	Precondition *Expr
	Signatories  *Expr
	Observers    *Expr
	Agreement    *Expr
	Choices      map[string]*Choice
	Location     *Location
	Key          *TemplateKey
}

type Choice struct {
	Name        string
	Consuming   bool
	Controllers *Expr
	Observers   *Expr
	SelfBinder  string
	ArgBinder   *VarWithType
	ReturnType  *Type
	Update      *Expr
	Location    *Location
	Authority   *Expr
}

type TemplateKey struct {
	Type        *Type
	Key         *Expr
	Maintainers *Expr
}

type Interface struct {
	Name         string
	Param        string
	Choices      map[string]*Choice
	Methods      map[string]*InterfaceMethod
	Precondition *Expr
	Location     *Location
}

type InterfaceMethod struct {
	Name     string
	Type     *Type
	Location *Location
}

type Value struct {
	Name     *ValueName
	Expr     *Expr
	IsTest   bool
	Location *Location
}

type ValueName struct {
	Module []string
	Name   string
}

type VarWithType struct {
	Var  string
	Type *Type
}

type Expr struct {
	Var                 *ExprVar
	Val                 *ExprVal
	Builtin             *ExprBuiltin
	PrimCon             *ExprPrimCon
	PrimLit             *ExprPrimLit
	RecCon              *ExprRecCon
	RecProj             *ExprRecProj
	RecUpd              *ExprRecUpd
	VariantCon          *ExprVariantCon
	EnumCon             *ExprEnumCon
	StructCon           *ExprStructCon
	StructProj          *ExprStructProj
	StructUpd           *ExprStructUpd
	App                 *ExprApp
	TyApp               *ExprTyApp
	Abs                 *ExprAbs
	TyAbs               *ExprTyAbs
	Case                *ExprCase
	Let                 *ExprLet
	Nil                 *ExprNil
	Cons                *ExprCons
	Update              *ExprUpdate
	Scenario            *ExprScenario
	OptionalNone        *ExprOptionalNone
	OptionalSome        *ExprOptionalSome
	ToAny               *ExprToAny
	FromAny             *ExprFromAny
	TypeRep             *ExprTypeRep
	ToAnyException      *ExprToAnyException
	FromAnyException    *ExprFromAnyException
	Throw               *ExprThrow
	ToInterface         *ExprToInterface
	FromInterface       *ExprFromInterface
	UnsafeFromInterface *ExprUnsafeFromInterface
	CallInterface       *ExprCallInterface
	SignatoryInterface  *ExprSignatoryInterface
	ObserverInterface   *ExprObserverInterface
	ViewInterface       *ExprViewInterface
	ChoiceController    *ExprChoiceController
	ChoiceObserver      *ExprChoiceObserver
	Experimental        *ExprExperimental
	Location            *Location
}

type ExprVar struct {
	Var string
}

type ExprVal struct {
	Ref *ValName
}

type ValName struct {
	Module []string
	Name   string
}

type ExprBuiltin struct {
	Builtin string
}

type ExprPrimCon struct {
	PrimCon string
}

type ExprPrimLit struct {
	PrimLit *PrimLit
}

type PrimLit struct {
	Int64        *int64
	Numeric      *string
	Text         *string
	Timestamp    *int64
	Party        *string
	Bool         *bool
	Unit         bool
	Date         *int32
	RoundingMode *int32
}

type ExprRecCon struct {
	Tycon  *TypeConName
	Fields []*FieldWithExpr
}

type FieldWithExpr struct {
	Field string
	Expr  *Expr
}

type TypeConName struct {
	Module []string
	Name   string
}

type ExprRecProj struct {
	Tycon  *TypeConName
	Field  string
	Record *Expr
}

type ExprRecUpd struct {
	Tycon  *TypeConName
	Field  string
	Record *Expr
	Update *Expr
}

type ExprVariantCon struct {
	Tycon   *TypeConName
	Variant string
	Arg     *Expr
}

type ExprEnumCon struct {
	Tycon *TypeConName
	Con   string
}

type ExprStructCon struct {
	Fields []*FieldWithExpr
}

type ExprStructProj struct {
	Field  string
	Struct *Expr
}

type ExprStructUpd struct {
	Field  string
	Struct *Expr
	Update *Expr
}

type ExprApp struct {
	Fun  *Expr
	Args []*Expr
}

type ExprTyApp struct {
	Expr  *Expr
	Types []*Type
}

type ExprAbs struct {
	Param []*VarWithType
	Body  *Expr
}

type ExprTyAbs struct {
	Param []*TypeVarWithKind
	Body  *Expr
}

type ExprCase struct {
	Scrut *Expr
	Alts  []*CaseAlt
}

type CaseAlt struct {
	Pattern *CasePattern
	Expr    *Expr
}

type CasePattern struct {
	Default      bool
	Variant      *CasePatternVariant
	Enum         *CasePatternEnum
	PrimCon      *CasePatternPrimCon
	Nil          bool
	Cons         *CasePatternCons
	OptionalNone bool
	OptionalSome *CasePatternOptionalSome
}

type CasePatternVariant struct {
	Con     *TypeConName
	Variant string
	Binder  string
}

type CasePatternEnum struct {
	Con         *TypeConName
	Constructor string
}

type CasePatternPrimCon struct {
	PrimCon string
}

type CasePatternCons struct {
	VarHead string
	VarTail string
}

type CasePatternOptionalSome struct {
	VarBody string
}

type ExprLet struct {
	Bindings []*Binding
	Body     *Expr
}

type Binding struct {
	Binder *VarWithType
	Bound  *Expr
}

type ExprNil struct {
	Type *Type
}

type ExprCons struct {
	Type  *Type
	Front []*Expr
	Tail  *Expr
}

type ExprUpdate struct {
	Update *Update
}

type Update struct {
	Pure              *UpdatePure
	Block             *UpdateBlock
	Create            *UpdateCreate
	CreateInterface   *UpdateCreateInterface
	Exercise          *UpdateExercise
	ExerciseInterface *UpdateExerciseInterface
	ExerciseByKey     *UpdateExerciseByKey
	Fetch             *UpdateFetch
	FetchInterface    *UpdateFetchInterface
	GetTime           *UpdateGetTime
	LookupByKey       *UpdateLookupByKey
	FetchByKey        *UpdateFetchByKey
	EmbedExpr         *UpdateEmbedExpr
	TryCatch          *UpdateTryCatch
}

type UpdatePure struct {
	Expr *Expr
}

type UpdateBlock struct {
	Bindings []*Binding
	Body     *Expr
}

type UpdateCreate struct {
	Template *TypeConName
	Expr     *Expr
}

type UpdateCreateInterface struct {
	Interface *TypeConName
	Expr      *Expr
}

type UpdateExercise struct {
	Template *TypeConName
	Choice   string
	Cid      *Expr
	Arg      *Expr
}

type UpdateExerciseInterface struct {
	Interface *TypeConName
	Choice    string
	Cid       *Expr
	Arg       *Expr
	Guard     *Expr
}

type UpdateExerciseByKey struct {
	Template *TypeConName
	Choice   string
	Key      *Expr
	Arg      *Expr
}

type UpdateFetch struct {
	Template *TypeConName
	Cid      *Expr
}

type UpdateFetchInterface struct {
	Interface *TypeConName
	Cid       *Expr
}

type UpdateGetTime struct{}

type UpdateLookupByKey struct {
	Template *TypeConName
	Key      *Expr
}

type UpdateFetchByKey struct {
	Template *TypeConName
	Key      *Expr
}

type UpdateEmbedExpr struct {
	Type *Type
	Body *Expr
}

type UpdateTryCatch struct {
	Type      *Type
	TryExpr   *Expr
	VarCatch  string
	CatchExpr *Expr
}

type ExprScenario struct {
	Scenario *Scenario
}

type Scenario struct {
	Pure       *ScenarioPure
	Block      *ScenarioBlock
	Commit     *ScenarioCommit
	MustFailAt *ScenarioMustFailAt
	Pass       *ScenarioPass
	GetTime    *ScenarioGetTime
	GetParty   *ScenarioGetParty
	EmbedExpr  *ScenarioEmbedExpr
}

type ScenarioPure struct {
	Expr *Expr
}

type ScenarioBlock struct {
	Bindings []*Binding
	Body     *Expr
}

type ScenarioCommit struct {
	Party   *Expr
	Update  *Expr
	RetType *Type
}

type ScenarioMustFailAt struct {
	Party   *Expr
	Update  *Expr
	RetType *Type
}

type ScenarioPass struct {
	Delta *Expr
}

type ScenarioGetTime struct{}

type ScenarioGetParty struct {
	Name *Expr
}

type ScenarioEmbedExpr struct {
	Type *Type
	Body *Expr
}

type ExprOptionalNone struct {
	Type *Type
}

type ExprOptionalSome struct {
	Type *Type
	Body *Expr
}

type ExprToAny struct {
	Type *Type
	Expr *Expr
}

type ExprFromAny struct {
	Type *Type
	Expr *Expr
}

type ExprTypeRep struct {
	Type *Type
}

type ExprToAnyException struct {
	Type *Type
	Expr *Expr
}

type ExprFromAnyException struct {
	Type *Type
	Expr *Expr
}

type ExprThrow struct {
	RetType       *Type
	ExceptionType *Type
	ExceptionExpr *Expr
}

type ExprToInterface struct {
	InterfaceType *TypeConName
	TemplateType  *TypeConName
	TemplateExpr  *Expr
}

type ExprFromInterface struct {
	InterfaceType *TypeConName
	TemplateType  *TypeConName
	InterfaceExpr *Expr
}

type ExprUnsafeFromInterface struct {
	InterfaceType  *TypeConName
	TemplateType   *TypeConName
	ContractIdExpr *Expr
	InterfaceExpr  *Expr
}

type ExprCallInterface struct {
	InterfaceType *TypeConName
	MethodName    string
	InterfaceExpr *Expr
}

type ExprSignatoryInterface struct {
	InterfaceType *TypeConName
	InterfaceExpr *Expr
}

type ExprObserverInterface struct {
	InterfaceType *TypeConName
	InterfaceExpr *Expr
}

type ExprViewInterface struct {
	InterfaceType *TypeConName
	InterfaceExpr *Expr
}

type ExprChoiceController struct {
	Template      *TypeConName
	Choice        string
	ContractExpr  *Expr
	ChoiceArgExpr *Expr
}

type ExprChoiceObserver struct {
	Template      *TypeConName
	Choice        string
	ContractExpr  *Expr
	ChoiceArgExpr *Expr
}

type ExprExperimental struct {
	Name string
	Type *Type
}

type Location struct {
	Module *ModuleRef
	Range  *Range
}

type ModuleRef struct {
	PackageRef *PackageRef
	ModuleName []string
}

type PackageRef struct {
	SumCase   string // "self" or "package_id"
	PackageId *string
}

type Range struct {
	StartLine int32
	StartCol  int32
	EndLine   int32
	EndCol    int32
}

type Metadata struct {
	Name         string
	Version      string
	Dependencies []string
	LangVersion  string
	CreatedBy    string
	SdkVersion   string
	CreatedAt    *time.Time
}
