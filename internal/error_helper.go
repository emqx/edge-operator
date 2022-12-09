package internal

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
)

// IsQuotaExceeded returns true if the error returned by the Kubernetes API is a forbidden error with the error message
// that the quota was exceeded
func IsQuotaExceeded(err error) bool {
	if k8sErrors.IsForbidden(err) {
		if strings.Contains(err.Error(), "exceeded quota") {
			return true
		}
	}
	return false
}
