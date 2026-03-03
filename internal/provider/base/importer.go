package base

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"strings"
)

// ImportIDWithSite parses import IDs in "site:id" format for the Plugin Framework.
func ImportIDWithSite(req resource.ImportStateRequest, resp *resource.ImportStateResponse) (string, string) {
	id := req.ID
	if id == "" {
		resp.Diagnostics.AddError("Invalid ID", "ID is required")
		return "", ""
	}

	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		if len(importParts) == 2 {
			return importParts[1], importParts[0]
		}
		resp.Diagnostics.AddError("Invalid ID", "ID contains too many colon-separated parts. Format should be 'site:id'")
		return "", ""
	}
	resp.Diagnostics.AddError("Invalid ID", "ID does not contain site part. Format should be 'site:id'")
	return "", ""
}
