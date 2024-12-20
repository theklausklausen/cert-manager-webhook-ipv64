package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/klog"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ipv64 "cert-manager-webhook-ipv64/internal/ipv64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var GroupName = os.Getenv("GROUP_NAME")
var Version = os.Getenv("VERSION")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName,
		&ipv64DNSProviderSolver{},
	)
}

type ipv64DNSProviderSolver struct {
	client *kubernetes.Clientset
}

type ipv64DNSProviderConfig struct {
	Email      string `json:"email"`
	SecretName string `json:"secretName"`
	Subdomain  string `json:"subdomain"`
}

func (c *ipv64DNSProviderSolver) Name() string {
	return "cert-manager-webhook-ipv64"
}

func (c *ipv64DNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	config, err := loadConfig(ch.Config)
	if err != nil {
		klog.Error("unable to load config: ", err)
		return err
	}

	if err != nil {
		klog.Error("unable to get secret: ", err)
		return fmt.Errorf("unable to get secret `%s`; %v", ch.ResourceNamespace, err)
	}

	ipv64Client, err := getClient(config, c.client, ch)
	if err != nil {
		klog.Error("unable to get client: ", err)
		return err
	}

	prefix := ch.ResolvedFQDN[:len(ch.ResolvedFQDN)-len(config.Subdomain)-1]

	err = ipv64Client.AddDNSRecord(
		config.Subdomain,
		prefix,
		ch.Key,
		"TXT")

	if err != nil {
		return fmt.Errorf("unable to add record `%s`; %v", ch.ResolvedFQDN, err)
	}

	klog.Info("Presented txt record ", ch.ResolvedFQDN)

	return nil
}

func (c *ipv64DNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	config, err := loadConfig(ch.Config)
	if err != nil {
		klog.Error("unable to load config: ", err)
		return err
	}

	if err != nil {
		klog.Error("unable to get secret: ", err)
		return fmt.Errorf("unable to get secret `%s`; %v", ch.ResourceNamespace, err)
	}

	ipv64Client, err := getClient(config, c.client, ch)
	if err != nil {
		klog.Error("unable to get client: ", err)
		return err
	}

	prefix := ch.ResolvedFQDN[:len(ch.ResolvedFQDN)-len(config.Subdomain)-1]

	err = ipv64Client.DeleteDNSRecord(
		config.Subdomain,
		prefix,
		ch.Key,
		"TXT")

	if err != nil {
		return fmt.Errorf("unable to delete record `%s`; %v", ch.ResolvedFQDN, err)
	}

	klog.Info("Deleted txt record ", ch.ResolvedFQDN)

	return nil
}

func (c *ipv64DNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	klog.Info("Starting cert-manager-webhook-ipv64...")
	klog.Info("Group name: ", GroupName)
	klog.Info("Solver name: ", c.Name())

	k8sClient, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	c.client = k8sClient

	return nil
}

func loadConfig(cfgJSON *extapi.JSON) (ipv64DNSProviderConfig, error) {
	cfg := ipv64DNSProviderConfig{}
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func getTokenFromSecret(secretName string, client *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (string, error) {
	sec, err := client.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		klog.Error("Error getting secret: ", err)
		return "", err
	}
	token, err := stringFromSecretData(sec.Data, "api-key")
	if err != nil {
		return "", fmt.Errorf("error decoding api-key: %v", err)
	}
	return token, nil
}

func getClient(cfg ipv64DNSProviderConfig, client *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (*ipv64.Client, error) {
	token, err := getTokenFromSecret(cfg.SecretName, client, ch)
	if err != nil {
		return nil, err
	}
	return ipv64.NewClient(token), nil
}

func stringFromSecretData(secretData map[string][]byte, key string) (string, error) {
	data, ok := secretData[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret data", key)
	}
	return string(data), nil
}
