package sendinbluev2

import (
	"context"
	"fmt"

	// "log"
	"regexp"
	"strings"

	"net/http"

	"github.com/trufflesecurity/trufflehog/pkg/common"
	"github.com/trufflesecurity/trufflehog/pkg/detectors"
	"github.com/trufflesecurity/trufflehog/pkg/pb/detectorspb"
)

type Scanner struct{}

// Ensure the Scanner satisfies the interface at compile time
var _ detectors.Detector = (*Scanner)(nil)

var (
	client = common.SaneHttpClient()

	keyPat = regexp.MustCompile(`\b(xkeysib\-[A-Za-z0-9_-]{81})\b`)
)

// Keywords are used for efficiently pre-filtering chunks.
// Use identifiers in the secret preferably, or the provider name.
func (s Scanner) Keywords() []string {
	return []string{"xkeysib"}
}

// FromData will find and optionally verify SendinBlueV2 secrets in a given set of bytes.
func (s Scanner) FromData(ctx context.Context, verify bool, data []byte) (results []detectors.Result, err error) {
	dataStr := string(data)

	matches := keyPat.FindAllStringSubmatch(dataStr, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		resMatch := strings.TrimSpace(match[1])

		s1 := detectors.Result{
			DetectorType: detectorspb.DetectorType_SendinBlueV2,
			Raw:          []byte(resMatch),
		}

		if verify {
			req, _ := http.NewRequest("GET", "https://api.sendinblue.com/v3/account", nil)
			req.Header.Add("api-key", fmt.Sprintf("%s", resMatch))
			res, err := client.Do(req)
			if err == nil {
				defer res.Body.Close()
				if res.StatusCode >= 200 && res.StatusCode < 300 {
					s1.Verified = true
				} else {
					if detectors.IsKnownFalsePositive(resMatch, detectors.DefaultFalsePositives, false) {
						continue
					}
				}
			}
		}

		results = append(results, s1)
	}

	return detectors.CleanResults(results), nil
}