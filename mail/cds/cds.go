package cds

type DocType string

const (
	UndefinedType DocType = "Undefined"
	DeclarationType = "Declaration"
	AmendmentType = "Amendment"
	CancellationType = "Cancellation"
	MovementType = "Movement"
)
