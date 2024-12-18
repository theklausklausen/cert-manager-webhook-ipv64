package ipv64

import (
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
	params.Add("add_record", subdomain)
	params.Add("praefix", praefix)
	params.Add("type", recordType)
	params.Add("content", content)

	req, err := http.NewRequest("POST", c.ApiUrl+"?"+params.Encode(), nil)
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

	klog.Info("Response Status: ", resp.Status, "Response Headers: ", resp.Header, "Response Body: ", string(body))

	klog.Info("Added record ", praefix, ".", subdomain)

	return nil
}

// func (e *ipv64DNSProviderSolver) getDNSRecord() error {
// 	// https://ipv64.net/api.php?get_domains

// 	params := url.Values{}
// 	params.Add("get_domains", e.domain)

// 	req, err := http.NewRequest("GET", "https://"+e.server+"/api.php?"+params.Encode(), nil)
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Authorization", "Bearer "+c.Token)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 	}

// 	var response interface{}
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		return err
// 	}

// 	var records []Record

// 	// itterate over each attribute of the json object
// 	for key, value := range response.(map[string]interface{}) {
// 		fmt.Println("Key:", key, "Value:", value)
// 		for _, _record := range value.([]interface{})["records"] {
// 			records = append(records, Record{
// 				RecordID: _record.RecordID,
// 				Content:  _record.Content,
// 				Type:     _record.Type,
// 				Praefix:  _record.Praefix,
// 				Domain:   key,
// 			})
// 		}
// 	}

// 	return nil
// }

// func (e *ipv64DNSProviderSolver) updateDNSRecord(q dns.Question, msg *dns.Msg, req *dns.Msg) error {

// 	return nil
// }

func (c *Client) DeleteDNSRecord(subdomain string, praefix string, content string, recordType string) error {

	klog.Info("call function DeleteDNSRecord: subdomain=", subdomain, ", praefix=", praefix, ", content=", content, ", recordType=", recordType)

	if recordType != "TXT" && recordType != "A" && recordType != "CNAME" && recordType != "MX" && recordType != "NS" && recordType != "PTR" && recordType != "SRV" && recordType != "SOA" && recordType != "AAAA" {
		return fmt.Errorf("unsupported record type: %s", recordType)
	}

	params := url.Values{}
	params.Add("del_record", subdomain)
	params.Add("praefix", praefix)
	params.Add("type", recordType)
	params.Add("content", content)

	req, err := http.NewRequest("DELETE", c.ApiUrl+"?"+params.Encode(), nil)
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

	klog.Info("Response Status: ", resp.Status, "Response Headers: ", resp.Header, "Response Body: ", resp.Body)

	klog.Info("Deleted record ", praefix, ".", subdomain)

	return nil
}
