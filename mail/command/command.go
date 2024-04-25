package command

import (
	"encoding/json"
	"encoding/xml"
	"mail/cds"
	"strings"
	"time"
	"fmt"
	_ "embed"
)

//go:embed Export.json
var exportJSON []byte

var UCRPrefix = "GB123456789012"

type Handler struct {
	Handler func(string)(string, error)
	DocType cds.DocType
}

var handlers = map[string]Handler{
	"EAC": Handler{Handler: processEAC, DocType: cds.MovementType},
	"CST": Handler{Handler: processCST, DocType: cds.MovementType},
	"EAA": Handler{Handler: processEAA, DocType: cds.MovementType},
	"EAL": Handler{Handler: processEAL, DocType: cds.MovementType},
	"EDL": Handler{Handler: processEDL, DocType: cds.MovementType},
	"EXP": Handler{Handler: processExport, DocType: cds.DeclarationType},
	"QUE": Handler{Handler: processQUE, DocType: cds.MovementType},
}

var emulatorRoutes map[string]string = map[string]string{
	"R1": "ttdocblockingF",
	"R2": "ttphysicalQ",
	"R3": "ttdocnonblockingU",
}

func Parse(content string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		handler, prs := handlers[strings.ToUpper(line)[0:3]]
		if prs {
			doc, err := handler.Handler(line)
			return doc, err
		}
	}
	return "", nil
}

func processQUE(command string) (string, error) {
	que := cds.MakeQuery()
	processQueryArgs(command, que)
	content, err := xml.MarshalIndent(que, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processEAC(command string) (string, error) {
	eac := cds.MakeConsolidation()
	processConsolidationArgs(command, eac)
	content, err := xml.MarshalIndent(eac, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processCST(command string) (string, error) {
	cst := cds.MakeShut()
	processConsolidationArgs(command, cst)
	content, err := xml.MarshalIndent(cst, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processEAA(command string) (string, error) {
	eaa := cds.MakeAnticipatedArrival()
	eaa.SetDUCR(makeDUCR())
	eaa.GoodsLocation = "GBAUDVRDOVDVR"
	processMovementArgs(command, eaa)
	content, err := xml.MarshalIndent(eaa, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processEAL(command string) (string, error) {
	eal := cds.MakeArrival()
	eal.SetDUCR(makeDUCR())
	eal.GoodsLocation = "GBAUDVRDOVDVR"
	eal.SetArrivalDate(time.Now())
	processMovementArgs(command, eal)
	content, err := xml.MarshalIndent(eal, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processEDL(command string) (string, error) {
	edl := cds.MakeDeparture()
	edl.SetDUCR(makeDUCR())
	edl.GoodsLocation = "GBAUDVRDOVDVR"
	edl.SetDepartureDate(time.Now())
	processMovementArgs(command, edl)
	content, err := xml.MarshalIndent(edl, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processExport(command string) (string, error) {
	exp := cds.MakePrelodgedExport()
	exp.FunctionalReferenceID = "CDSDEC" + makeUniqueString13()
	if err := json.Unmarshal(exportJSON, exp); err != nil {
		return "", err
	}
	processDecArgs(command, exp)
	exp.UpdateComputedTotals()
	if len(exp.GoodsShipment.PreviousDocuments) == 0 {
		exp.SetDUCR(makeDUCR())
	}
	content, err := xml.MarshalIndent(cds.WrapMetaData(exp), "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func makeUniqueString13() string {
	return strings.Replace(time.Now().UTC().Format("0201150405.000"), ".", "", 1)
}

func makeDUCR() string {
	lastDigitOfYear := time.Now().String()[3:4]
	return lastDigitOfYear + UCRPrefix + "-CDSDEC" + makeUniqueString13()
}

func makeMUCR() string {
	return "GB/" + UCRPrefix[2:] + "-CDSDEC" + makeUniqueString13()
}

func isRoute(arg string) bool {
	arg = strings.ToUpper(arg)
	_, present := emulatorRoutes[arg]
	return present
}

func processDecArgs(args string, dec *cds.Declaration) {
	waitNext := 0
	for _, arg := range strings.Split(args, " ")[1:] {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		argU := strings.ToUpper(arg)
		if waitNext == 1 {
			dec.SetGoodsDescription(arg)
			waitNext = 0
		} else if waitNext == 2 {
			dec.SetPackagingType(arg)
			waitNext = 0
		} else if cds.IsDUCR(arg) {
			dec.SetDUCR(arg)
		} else if cds.IsMUCR(arg) {
			dec.SetMUCR(arg)
		} else if isRoute(arg) {
			desc, _ := emulatorRoutes[argU]
			dec.SetGoodsDescription(desc)
		} else if argU == "DESC" {
			waitNext = 1
		} else if argU == "PKTYPE" {
			waitNext = 2
		}
	}
}

func processMovementArgs(args string, movement *cds.Movement) {
	waitNext := 0
	for _, arg := range strings.Split(strings.ToUpper(args), " ")[1:] {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		if waitNext == 1 {
			movement.GoodsLocation = arg
			waitNext = 0
		} else if waitNext == 2 {
			movement.MasterOpt = strings.ToUpper(arg)
			waitNext = 0
		} else if cds.IsDUCR(arg) {
			movement.SetDUCR(arg)
		} else if cds.IsMUCR(arg) {
			movement.SetMUCR(arg)
		} else if arg == "MUCR" {
			movement.SetMUCR(makeMUCR())
		} else if arg == "LOC" {
			waitNext = 1
		} else if arg == "MASTEROPT" {
			waitNext = 2
		}
	}
}

func processQueryArgs(args string, query *cds.Query) {
	for _, arg := range strings.Split(strings.ToUpper(args), " ")[1:] {
		if strings.HasPrefix(arg, "<") {
			arg = arg[1:]
		}
		if strings.HasSuffix(arg, ">") {
			arg = arg[:len(arg) - 1]
		}
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		if cds.IsDUCR(arg) {
			query.SetDUCR(arg)
		} else if cds.IsMUCR(arg) {
			query.SetMUCR(arg)
		}
	}
}

func processConsolidationArgs(args string, c *cds.Consolidation) {
	for _, arg := range strings.Split(strings.ToUpper(args), " ")[1:] {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		fmt.Printf("%s: isMUCR? %v\n", arg, cds.IsMUCR(arg))
		fmt.Printf("%s: isDUCR? %v\n", arg, cds.IsDUCR(arg))
		if cds.IsDUCR(arg) {
			c.SetChild(arg, "D")
		} else if cds.IsMUCR(arg) {
			if c.MasterUCR != "" {
				c.SetChild(arg, "M")
			} else {
				c.MasterUCR = arg
			}
			fmt.Printf("---> %v\n", *c)
		} else if arg == "MUCR" {
			if c.MasterUCR != "" {
				c.SetChild(makeMUCR(), "M")
			} else {
				c.MasterUCR = makeMUCR()
			}
		}
	}
}
