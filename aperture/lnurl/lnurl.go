package lnurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Lnurl struct {
	Lud16 string
	Lnurl string
}

// {"status":"OK","successAction":{"tag":"message","message":"Thanks, sats received!"},"verify":"https://getalby.com/lnurlp/moti/verify/v5bzeXMWXzFRPKoSJTqYe6ZX","routes":[],"pr":"lnbc10n1pjlqd2spp5ejf08qxta88prqvm9q7j4e6g2jzt0ux00spqnh39t7wwdp0z69cqhp5cezvxddw0lgesz3xpr67q7v8tux7uv5h5vdwukrlgg3m22ce6dcscqzzsxqyz5vqsp5ujym2lynsdhda5znuk8h0wm7kky930ty9pxl6aktfffgue4x5upq9qyyssqgtk0wr34n2jnmnv4d3lqlmdvrqz3ekme5s2r3vhr5kqh4rxj6rl3vg4t3ppygvl9ymg28f5pg9etv6zysuvy3jcagetcvfryjhv04jspul97hy"}
type LnurlResponse struct {
	Status        string `json:"status"`
	SuccessAction struct {
		Tag     string `json:"tag"`
		Message string `json:"message"`
	} `json:"successAction"`
	Verify string `json:"verify"`
	Routes []struct {
		Pubkey string `json:"pubkey"`
		Alias  string `json:"alias"`
	} `json:"routes"`
	Pr string `json:"pr"`
}

func NewLnurl(lud16 string) (*Lnurl, error) {
	parts := strings.Split(lud16, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid lud16 format")
	}
	name, domain := parts[0], parts[1]
	return &Lnurl{
		Lud16: lud16,
		Lnurl: fmt.Sprintf("https://%s/.well-known/lnurlp/%s", domain, name),
	}, nil
}

func (l *Lnurl) GetInvoice(price int64) (string, error) {
	resp, err := http.Get(l.Lnurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body LnurlResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("call zap endpoint error: %s", resp.Status)
	}
	if body.Pr == "" {
		return "", fmt.Errorf("invoice not found")
	}

	return body.Pr, nil
}
