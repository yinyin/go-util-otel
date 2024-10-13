package utilotel

import (
	"errors"
	"fmt"
)

var ErrUnknownExportMode = errors.New("unknown export mode")

const switchModeCount = 3

type ExportMode int

const (
	ExportModeNoop ExportMode = iota
	ExportModeSTDOUT
	ExportModeOTLPgRPC
	exportModeBoundary
)

func checkExportMode(mode ExportMode) error {
	if (mode < 0) || (mode >= exportModeBoundary) {
		return fmt.Errorf("%w: %d", ErrUnknownExportMode, mode)
	}
	return nil
}
