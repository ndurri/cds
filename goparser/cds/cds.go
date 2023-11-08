package cds

type DocType string

const (
	UndefinedType DocType = "Undefined"
	MovementType = "Movement"
	DeclarationType = "Declaration"
	AmendmentType = "Amendment"
	CancellationType = "Cancellation"
	QueryType = "Query"
)

type Request interface {
	DocType() DocType
}

type Consolidation struct {}
type Movement struct {}
type Query struct {}

type MetaData struct {
	Declaration *Declaration
}

type Declaration struct {
	FunctionCode int
	TypeCode string
}

func (r *Consolidation) DocType() DocType {return MovementType}
func (r *Movement) DocType() DocType {return MovementType}
func (r *Query) DocType() DocType {return MovementType}
func (r *MetaData) DocType() DocType {
	if r.Declaration.FunctionCode == 9 {
		return DeclarationType
	} else if r.Declaration.FunctionCode == 13 {
		if r.Declaration.TypeCode == "COR" {
			return AmendmentType
		} else if r.Declaration.TypeCode == "INV" {
			return CancellationType
		}
	}
	return UndefinedType
}
