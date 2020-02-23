package coap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-ocf/go-coap"
	"github.com/pion/dtls/v2"
)

const (
	RootDevices = 15001
	RootGroups  = 15004
)

type Client struct {
	conn *coap.ClientConn
}

func NewClient(network, address, username, psk string) (*Client, error) {
	conn, err := coap.DialDTLSWithTimeout(network, address, &dtls.Config{
		PSK:             func(hint []byte) ([]byte, error) { return []byte(psk), nil },
		PSKIdentityHint: []byte(username),
		CipherSuites: []dtls.CipherSuiteID{
			dtls.TLS_PSK_WITH_AES_128_CCM,
			dtls.TLS_PSK_WITH_AES_128_CCM_8,
			dtls.TLS_PSK_WITH_AES_128_GCM_SHA256,
		},
	}, 3*time.Second)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) Auth(username string) (psk string, err error) {
	buf, err := json.Marshal(struct {
		Username string `json:"9090"`
	}{Username: username})
	if err != nil {
		return "", fmt.Errorf("error marshaling request payload: %w", err)
	}

	msg, err := c.conn.Post("/15011/9063", coap.AppJSON, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}

	if msg.Code() > 100 {
		return "", fmt.Errorf("response code %d (%s)", msg.Code(), msg.Code().String())
	}

	var response struct {
		PreSharedKey    string `json:"9091"`
		FirmwareVersion string `json:"9029"`
	}
	if err := json.Unmarshal(msg.Payload(), &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response payload: %w", err)
	}

	return response.PreSharedKey, nil
}

func (c *Client) GetDevice(id int) (d Device, err error) {
	err = c.get(fmt.Sprintf("/15001/%d", id), &d)
	return d, err
}

func (c *Client) ListDevices() ([]Device, error) {
	var ids []int
	if err := c.get("/15001", &ids); err != nil {
		return nil, fmt.Errorf("error listing device IDs: %w", err)
	}

	var devices []Device
	for _, id := range ids {
		d, err := c.GetDevice(id)
		if err != nil {
			return nil, fmt.Errorf("error getting device %d: %w", id, err)
		}
		devices = append(devices, d)
	}

	return devices, nil
}

func (c *Client) GetGroup(id int) (g Group, err error) {
	err = c.get(fmt.Sprintf("/15004/%d", id), &g)
	return g, err
}

func (c *Client) ListGroups() ([]Group, error) {
	var ids []int
	if err := c.get("/15004", &ids); err != nil {
		return nil, fmt.Errorf("error listing group IDs: %w", err)
	}

	var groups []Group
	for _, id := range ids {
		g, err := c.GetGroup(id)
		if err != nil {
			return nil, fmt.Errorf("error getting group %d: %w", id, err)
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (c *Client) SetLightControlState(root, id int, on bool) error {
	var st int
	if on {
		st = 1
	}
	return c.put(fmt.Sprintf("/%d/%d", root, id), struct {
		State int `json:"5850"`
	}{
		State: st,
	})
}

func (c *Client) SetLightControlDimmer(root, id int, dimmer int, transition time.Duration) error {
	return c.put(fmt.Sprintf("/%d/%d", root, id), struct {
		Dimmer     int `json:"5851"` // 0..255
		Transition int `json:"5712"` // tenths of a second
	}{
		Dimmer:     dimmer,
		Transition: int(transition.Seconds() * 10),
	})
}

func (c *Client) SetLightControlMireds(root, id int, mireds int, transition time.Duration) error {
	return c.put(fmt.Sprintf("/%d/%d", root, id), struct {
		Mireds     int `json:"5711"` // 250..454
		Transition int `json:"5712"` // tenths of a second
	}{
		Mireds:     mireds,
		Transition: int(transition.Seconds() * 10),
	})
}

func (c *Client) get(path string, response interface{}) error {
	msg, err := c.conn.Get(path)
	if err != nil {
		return fmt.Errorf("error making Get request: %w", err)
	}

	if msg.Code() > 100 {
		return fmt.Errorf("response code %d (%s)", msg.Code(), msg.Code().String())
	}

	fmt.Fprintf(os.Stderr, "### get(%s): %s\n", path, string(msg.Payload()))

	if err := json.Unmarshal(msg.Payload(), response); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

func (c *Client) put(path string, request interface{}) error {
	buf, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	msg, err := c.conn.Put(path, coap.AppJSON, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("error making Put request: %w", err)
	}

	if msg.Code() > 100 {
		return fmt.Errorf("response code %d (%s)", msg.Code(), msg.Code().String())
	}

	return nil
}

//
//
//

type Resource struct {
	Name      string    `json:"9001"`
	CreatedAt Timestamp `json:"9002"`
	ID        int       `json:"9003"`
}

//
//
//

type LightControl struct {
	State         OnOff      `json:"5850"`
	Dimmer        Percent255 `json:"5851"`
	LightColorHex string     `json:"5706"`
	LightColorX   int        `json:"5709"`
	LightColorY   int        `json:"5710"`
	LightMireds   int        `json:"5711"`
	Unknown       int        `json:"5717"`
}

type LightControlInput struct {
	State         *OnOff      `json:"5850,omitempty"`
	Dimmer        *Percent255 `json:"5851,omitempty"`
	LightColorHex *string     `json:"5706,omitempty"`
	LightColorX   *int        `json:"5709,omitempty"`
	LightColorY   *int        `json:"5710,omitempty"`
	LightMireds   *int        `json:"5711,omitempty"`
	Transition    *int        `json:"5712,omitempty"`
}

//
//
//

type Device struct {
	Resource
	DeviceInfo struct {
		Manufacturer string      `json:"0"`
		Model        string      `json:"1"`
		Serial       string      `json:"2"`
		Firmware     string      `json:"3"`
		PowerSource  PowerSource `json:"6"`
		BatteryLevel Percent100  `json:"9"`
	} `json:"3"`
	LastSeen     Timestamp      `json:"9020"`
	Reachable    YesNo          `json:"9019"`
	LightControl []LightControl `json:"3311"`
}

func (d Device) Short() string {
	return fmt.Sprintf("%d: %s (%s)", d.ID, d.Name, d.DeviceInfo.Model)
}

func (d Device) Long() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", d.Name)
	fmt.Fprintf(&b, "Created at: %s\n", d.CreatedAt)
	fmt.Fprintf(&b, "ID: %d\n", d.ID)
	fmt.Fprintf(&b, "Manufacturer: %s\n", d.DeviceInfo.Manufacturer)
	fmt.Fprintf(&b, "Model: %s\n", d.DeviceInfo.Model)
	fmt.Fprintf(&b, "Serial: %s\n", d.DeviceInfo.Serial)
	fmt.Fprintf(&b, "Firmware: %s\n", d.DeviceInfo.Firmware)
	fmt.Fprintf(&b, "Power source: %s\n", d.DeviceInfo.PowerSource)
	fmt.Fprintf(&b, "Battery level: %d\n", d.DeviceInfo.BatteryLevel)
	fmt.Fprintf(&b, "Last seen: %s\n", d.LastSeen)
	fmt.Fprintf(&b, "Reachable: %s\n", d.Reachable)
	fmt.Fprintf(&b, "Light control count: %d\n", len(d.LightControl))
	for i, c := range d.LightControl {
		fmt.Fprintf(&b, "Light control %d: State: %s\n", i+1, c.State)
		fmt.Fprintf(&b, "Light control %d: Dimmer: %s\n", i+1, c.Dimmer)
		fmt.Fprintf(&b, "Light control %d: Light color (hex): %s\n", i+1, c.LightColorHex)
		fmt.Fprintf(&b, "Light control %d: Light color (X): %d\n", i+1, c.LightColorX)
		fmt.Fprintf(&b, "Light control %d: Light color (Y): %d\n", i+1, c.LightColorY)
		fmt.Fprintf(&b, "Light control %d: Light mireds: %d\n", i+1, c.LightMireds)
	}
	return strings.TrimSpace(b.String())
}

//
//
//

type Group struct {
	Resource
	State         OnOff      `json:"5850"`
	Dimmer        Percent255 `json:"5851"`
	LightColorHex string     `json:"5706"`
	MoodID        int        `json:"9039"`
	GroupMembers  struct {
		HSLink struct {
			IDs []int `json:"9003"`
		} `json:"15002"`
	} `json:"9018"`
}

func (g Group) Short() string {
	n := len(g.GroupMembers.HSLink.IDs)
	return fmt.Sprintf("%d: %s (%s) - %d member%s", g.ID, g.Name, g.State, n, plural(n))
}

func (g Group) Long() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", g.Name)
	fmt.Fprintf(&b, "Created at: %s\n", g.CreatedAt)
	fmt.Fprintf(&b, "ID: %d\n", g.ID)
	fmt.Fprintf(&b, "State: %s\n", g.State)
	fmt.Fprintf(&b, "Dimmer: %s\n", g.Dimmer)
	fmt.Fprintf(&b, "Light color: %s\n", g.LightColorHex)
	fmt.Fprintf(&b, "Mood ID: %d\n", g.MoodID)
	fmt.Fprintf(&b, "Member count: %d\n", len(g.GroupMembers.HSLink.IDs))
	for i, id := range g.GroupMembers.HSLink.IDs {
		fmt.Fprintf(&b, "Member %d: %d\n", i+1, id)
	}
	return strings.TrimSpace(b.String())
}

type SetGroupProperties struct {
	State       int
	Dimmer      int
	ColorHex    string
	ColorX      int
	ColorY      int
	ColorMireds int
	Transition  int
}

//
//
//

type Percent255 int

func (p Percent255) String() string {
	return fmt.Sprintf("%d%%", int(100*(float64(p)/float64(255))))
}

type Percent100 int

func (p Percent100) String() string {
	return fmt.Sprintf("%d%%", p)
}

type OnOff int

func (o OnOff) String() string {
	if o == 0 {
		return "off"
	}
	return "on"
}

type YesNo int

func (yn YesNo) String() string {
	if yn == 0 {
		return "no"
	}
	return "yes"
}

type Timestamp int

func (ts Timestamp) String() string {
	t := time.Unix(int64(ts), 0)
	return fmt.Sprintf("%s (%s ago)", t.Format(time.RubyDate), since(t))
}

type PowerSource int

func (ps PowerSource) String() string {
	switch ps {
	case 1:
		return "internal battery"
	case 2:
		return "external battery"
	case 3:
		return "battery"
	case 4:
		return "power over ethernet"
	case 5:
		return "USB"
	case 6:
		return "mains"
	case 7:
		return "solar"
	default:
		return fmt.Sprintf("unknown power source (%d)", ps)
	}
}

//
//
//

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func since(t time.Time) string {
	d := time.Since(t)
	switch {
	case d > 24*time.Hour:
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	case d > time.Hour:
		return fmt.Sprintf("%dh", d/(time.Hour))
	case d > time.Minute:
		return fmt.Sprintf("%dm", d/(time.Second))
	default:
		return d.Truncate(time.Second).String()
	}
}
