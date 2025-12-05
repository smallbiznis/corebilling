package quota

import (
	"errors"

	"github.com/smallbiznis/corebilling/internal/quota/repository"
)

type QuotaLimit = repository.QuotaLimit
type QuotaUsage = repository.QuotaUsage
type UpsertQuotaUsageParams = repository.UpsertQuotaUsageParams

var (
	ErrQuotaSoftWarning = errors.New("quota soft warning")
	ErrQuotaExceeded    = errors.New("quota exceeded")
)
