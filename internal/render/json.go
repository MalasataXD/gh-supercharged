package render

import (
	"encoding/json"
	"fmt"
	"os"
)

func JSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("render json: %w", err)
	}
	return nil
}
