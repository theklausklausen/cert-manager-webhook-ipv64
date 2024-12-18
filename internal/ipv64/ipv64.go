package ipv64

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"k8s.io/klog"
)

const ApiUrl = "https://ipv64.net/api.php"

var client Client

type Record struct {
	RecordID int
	Content  string
	Type     string
	Praefix  string
	Domain   string
}

type Client struct {
	ApiUrl string
	Token  string
}

func NewClient(token string) *Client {
	klog.Info("create new ipv64 client")
	if client == (Client{}) {
		client = Client{
			ApiUrl: ApiUrl,
			Token:  token,
		}
	}
	return &client
}

func (c *Client) AddDNSRecord(subdomain string, praefix string, content string, recordType string) error {

	// klog.Info("call function AddDNSRecord: subdomain=%s, praefix=%s, content=%s, recordType=%s")
	klog.Info("call function AddDNSRecord: subdomain=", subdomain, ", praefix=", praefix, ", content=", content, ", recordType=", recordType)

	if recordType != "TXT" && recordType != "A" && recordType != "CNAME" && recordType != "MX" && recordType != "NS" && recordType != "PTR" && recordType != "SRV" && recordType != "SOA" && recordType != "AAAA" {
		klog.Error("unsupported record type: ", recordType)
		return fmt.Errorf("unsupported record type: %s", recordType)
	}

	// params := url.Values{
	// 	"add_record": {subdomain},
	// 	"praefix":    {praefix},
	// 	"type":       {recordType},
	// 	"content":    {content},
	// }
	// encodedParams := params.Encode()

	data := bytes.Buffer{}
	data.Write([]byte("add_record="))
	data.Write([]byte(subdomain))
	data.Write([]byte("&praefix="))
	data.Write([]byte(praefix))
	data.Write([]byte("&type="))
	data.Write([]byte(recordType))
	data.Write([]byte("&content="))
	data.Write([]byte(content))

	klog.Info("URL: ", url)

	req, err := http.NewRequest("POST", c.ApiUrl, data.Bytes())
	if err != nil {
		klog.Error("error creating request: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.Error("error sending request: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		klog.Error("unexpected status code: ", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	klog.Infoln("Response Status: ", resp.Status)
	klog.Infoln("Response Headers: ", resp.Header)
	klog.Infoln("Response Body: ", string(body))

	klog.Info("Added record ", praefix, ".", subdomain)

	return nil
}

func (c *Client) DeleteDNSRecord(subdomain string, praefix string, content string, recordType string) error {

	klog.Info("call function DeleteDNSRecord: subdomain=", subdomain, ", praefix=", praefix, ", content=", content, ", recordType=", recordType)

	if recordType != "TXT" && recordType != "A" && recordType != "CNAME" && recordType != "MX" && recordType != "NS" && recordType != "PTR" && recordType != "SRV" && recordType != "SOA" && recordType != "AAAA" {
		return fmt.Errorf("unsupported record type: %s", recordType)
	}

	// params := url.Values{
	// 	"del_record": {subdomain},
	// 	"praefix":    {praefix},
	// 	"type":       {recordType},
	// 	"content":    {content},
	// }
	// encodedParams := params.Encode()

	data := bytes.Buffer{}
	data.Write([]byte("del_record="))
	data.Write([]byte(subdomain))
	data.Write([]byte("&praefix="))
	data.Write([]byte(praefix))
	data.Write([]byte("&type="))
	data.Write([]byte(recordType))
	data.Write([]byte("&content="))
	data.Write([]byte(content))

	req, err := http.NewRequest("DELETE", c.ApiUrl, data.Bytes())
	if err != nil {
		klog.Error("error creating request: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.Error("error sending request: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		klog.Error("unexpected status code: ", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	klog.Infoln("Response Status: ", resp.Status)
	klog.Infoln("Response Headers: ", resp.Header)
	klog.Infoln("Response Body: ", string(body))

	klog.Info("Added record ", praefix, ".", subdomain)

	return nil
}
