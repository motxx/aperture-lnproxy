package lnurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Lnurl struct {
	Lud16 string
	Lnurl string
}

// {"status":"OK","tag":"payRequest","commentAllowed":255,"callback":"https://getalby.com/lnurlp/moti/callback","metadata":"[[\"text/identifier\",\"moti@getalby.com\"],[\"text/plain\",\"Sats for moti\"]]","minSendable":1000,"maxSendable":500000000,"payerData":{"name":{"mandatory":false},"email":{"mandatory":false},"pubkey":{"mandatory":false}},"nostrPubkey":"79f00d3f5a19ec806189fcab03c1be4ff81d18ee4f653c88fac41fe03570f432","allowsNostr":true}%
type LnurlResponse struct {
	Status      string `json:"status"`
	Tag         string `json:"tag"`
	Comment     int    `json:"commentAllowed"`
	Callback    string `json:"callback"`
	Metadata    string `json:"metadata"`
	MinSendable int64  `json:"minSendable"`
	MaxSendable int64  `json:"maxSendable"`
	PayerData   struct {
		Name struct {
			Mandatory bool `json:"mandatory"`
		} `json:"name"`
		Email struct {
			Mandatory bool `json:"mandatory"`
		} `json:"email"`
		Pubkey struct {
			Mandatory bool `json:"mandatory"`
		} `json:"pubkey"`
	} `json:"payerData"`
	NostrPubkey string `json:"nostrPubkey"`
	AllowsNostr bool   `json:"allowsNostr"`
}

// {"status":"OK","successAction":{"tag":"message","message":"Thanks, sats received!"},"verify":"https://getalby.com/lnurlp/moti/verify/v5bzeXMWXzFRPKoSJTqYe6ZX","routes":[],"pr":"lnbc10n1pjlqd2spp5ejf08qxta88prqvm9q7j4e6g2jzt0ux00spqnh39t7wwdp0z69cqhp5cezvxddw0lgesz3xpr67q7v8tux7uv5h5vdwukrlgg3m22ce6dcscqzzsxqyz5vqsp5ujym2lynsdhda5znuk8h0wm7kky930ty9pxl6aktfffgue4x5upq9qyyssqgtk0wr34n2jnmnv4d3lqlmdvrqz3ekme5s2r3vhr5kqh4rxj6rl3vg4t3ppygvl9ymg28f5pg9etv6zysuvy3jcagetcvfryjhv04jspul97hy"}
type LnurlCallbackResponse struct {
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
	Pr     string `json:"pr"`
	Reason string `json:"reason"`
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

func (l *Lnurl) GetInvoice(amount_sats int64) (string, error) {
	resp, err := http.Get(l.Lnurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body LnurlResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK || body.Status != "OK" {
		res, _ := json.Marshal(resp.Body)
		return "", fmt.Errorf("request zap endpoint error. http status: %s, body: %s", resp.Status, string(res))
	}
	if body.Callback == "" {
		return "", fmt.Errorf("callback not found")
	}

	u, err := url.Parse(body.Callback)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("amount", fmt.Sprintf("%d", amount_sats*1000))
	u.RawQuery = q.Encode()

	resp, err = http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var callbackBody LnurlCallbackResponse
	if err := json.NewDecoder(resp.Body).Decode(&callbackBody); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK || callbackBody.Status != "OK" {
		res, _ := json.Marshal(callbackBody)
		return "", fmt.Errorf("request zap endpoint callback error. http status: %s, body: %s", resp.Status, string(res))
	}
	if callbackBody.Pr == "" {
		return "", fmt.Errorf("invoice not found")
	}

	return callbackBody.Pr, nil
}
