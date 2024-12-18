package ipv64

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

type AddRecordResponse struct {
	Info      string `json:"info"`
	Status    string `json:"status"`
	AddRecord string `json:"add_record"`
}

type DeleteRecordResponse struct {
	Info      string `json:"info"`
	Status    string `json:"status"`
	DelRecord string `json:"del_record"`
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

	params := url.Values{}
	params.Set("add_record", subdomain)
	params.Set("praefix", praefix)
	params.Set("type", recordType)
	params.Set("content", content)

	klog.Info("params: ", params)
	klog.Info("encoded params: ", params.Encode())

	req, err := http.NewRequest("POST", c.ApiUrl, bytes.NewBufferString(params.Encode()))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	response := AddRecordResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		klog.Error("error unmarshalling response body: ", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest && response.AddRecord == "dns record already there" {
			klog.Warningln("DNS record already there")
			return nil
		}
		klog.Error("Could not add record: ", response.Info)
		klog.V(4).Infoln("Response: ", response)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	klog.V(4).Infoln("Response: ", response)

	klog.Info("Added record ", praefix, ".", subdomain)

	return nil
}

func (c *Client) DeleteDNSRecord(subdomain string, praefix string, content string, recordType string) error {

	klog.Info("call function DeleteDNSRecord: subdomain=", subdomain, ", praefix=", praefix, ", content=", content, ", recordType=", recordType)

	if recordType != "TXT" && recordType != "A" && recordType != "CNAME" && recordType != "MX" && recordType != "NS" && recordType != "PTR" && recordType != "SRV" && recordType != "SOA" && recordType != "AAAA" {
		return fmt.Errorf("unsupported record type: %s", recordType)
	}

	params := url.Values{}
	params.Set("del_record", subdomain)
	params.Set("praefix", praefix)
	params.Set("type", recordType)
	params.Set("content", content)
	// encodedParams := params.Encode()

	req, err := http.NewRequest("DELETE", c.ApiUrl, bytes.NewBufferString(params.Encode()))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error("error reading response body: ", err)
		return err
	}

	response := DeleteRecordResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		klog.Error("error unmarshalling response body: ", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusAccepted && response.DelRecord == "del_record" {
			klog.Info("Deleted record ", praefix, ".", subdomain)
			return nil
		}
		klog.Error("Could not delete record: ", response.Info)
		klog.V(4).Infoln("Response: ", response)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	klog.V(4).Infoln("Response: ", response)

	klog.Info("Deleted record ", praefix, ".", subdomain)

	return nil
}
