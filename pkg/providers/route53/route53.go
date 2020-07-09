package route53

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/go-ini/ini"
	"github.com/greenpau/dyndns/pkg/record"
	"github.com/greenpau/dyndns/pkg/utils"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RegistrationProvider is a controller for updating DNS records hosted byo
// AWS Route 53 service.
type RegistrationProvider struct {
	Provider        string `json:"type" yaml:"type"`
	ZoneID          string `json:"zone_id" yaml:"zone_id"`
	Credentials     string `json:"credentials" yaml:"credentials"`
	ProfileName     string `json:"profile_name" yaml:"profile_name"`
	accessKeyID     string
	secretAccessKey string
	region          string
	log             *zap.Logger
}

// Validate validates an instance op *RegistrationProvider.
func (p *RegistrationProvider) Validate() error {
	if p.ZoneID == "" {
		return fmt.Errorf("provider requires a hosted zone id")
	}
	if p.Credentials == "" {
		return fmt.Errorf("aws credentials not found")
	}
	if p.Provider != "route53" {
		return fmt.Errorf("provider mismatch: %s (config) vs. route53 (expected)", p.Provider)
	}
	return nil
}

func (p *RegistrationProvider) loadCredentials() error {
	if p.Credentials[0] == '~' {
		hd, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to load credentials: %s", err)
		}
		p.Credentials = filepath.Join(hd, p.Credentials[1:])
	}
	cfg, err := ini.Load(p.Credentials)
	if err != nil {
		return fmt.Errorf("failed to load credentials from %s: %s", p.Credentials, err)
	}

	section := cfg.Section(p.ProfileName)
	if section == nil {
		return fmt.Errorf("failed to load profile %s from %s", p.ProfileName, p.Credentials)
	}

	p.accessKeyID = section.Key("aws_access_key_id").String()
	p.secretAccessKey = section.Key("aws_secret_access_key").String()
	p.region = section.Key("region").String()

	if p.accessKeyID == "" {
		return fmt.Errorf(
			"failed to load aws_access_key_id from profile %s in %s",
			p.ProfileName, p.Credentials,
		)
	}

	if p.secretAccessKey == "" {
		return fmt.Errorf(
			"failed to load aws_secret_access_key from profile %s in %s",
			p.ProfileName, p.Credentials,
		)
	}

	if p.region == "" {
		p.region = "us-east-1"
	}

	return nil
}

// Configure configures  an instance op *RegistrationProvider.
func (p *RegistrationProvider) Configure(logger *zap.Logger) error {
	p.log = logger
	if p.ProfileName == "" {
		p.ProfileName = "default"
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := p.loadCredentials(); err != nil {
		return err
	}
	p.log.Debug(
		"found aws credentials",
		zap.String("aws_access_key_id", utils.MaskSecret(p.accessKeyID, 4, 4)),
		zap.String("region", p.region),
		zap.String("aws_secret_access_key", utils.MaskSecret(p.secretAccessKey, 4, 4)),
	)
	return nil
}

// GetProvider returns the provider name associated with RegistrationProvider.
func (p *RegistrationProvider) GetProvider() string {
	return p.Provider
}

// Register registers a record with RegistrationProvider.
func (p *RegistrationProvider) Register(r *record.RegistrationRecord) error {
	if r.Name == "" {
		return fmt.Errorf("record name is empty")
	}
	nameParts := strings.SplitN(r.Name, ".", 2)
	hostname := nameParts[0]
	expDomain := nameParts[1]
	fqdn := r.Name
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	ip4, err := r.GetAddress(4)
	if err != nil {
		return err
	}

	p.log.Debug(
		"received registration request",
		zap.Any("record", r),
		zap.Any("address", ip4),
	)

	// Acquire AWS Session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(p.region),
		Credentials: credentials.NewStaticCredentials(p.accessKeyID, p.secretAccessKey, ""),
	})
	if err != nil {
		return fmt.Errorf("failed create aws session: %s", err)
	}

	// Connect to AWS Service
	svc := route53.New(sess)

	// Get information about Route 53 Zone
	hostedZoneRequest := &route53.GetHostedZoneInput{
		Id: aws.String(p.ZoneID),
	}
	hostedZoneResponse, err := svc.GetHostedZone(hostedZoneRequest)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				return fmt.Errorf("zone id %s not found: %s", p.ZoneID, aerr.Error())
			case route53.ErrCodeInvalidInput:
				return fmt.Errorf("invalid get hosted zone request in zone id %s: %s", p.ZoneID, aerr.Error())
			default:
				return fmt.Errorf("get hosted zone request failed: %s", aerr.Error())
			}
		}
		return fmt.Errorf("get hosted zone request failed: %s", err.Error())
	}

	if hostedZoneResponse.HostedZone == nil {
		return fmt.Errorf("get hosted zone request returned nil")
	}

	domain := strings.TrimRight(*hostedZoneResponse.HostedZone.Name, ".")
	if expDomain != domain {
		return fmt.Errorf("hosted zone mismatch: %s (expected) vs. %s (actual)", expDomain, domain)
	}

	var reverseDomain string
	domainParts := strings.Split(domain, ".")
	for i := len(domainParts) - 1; i >= 0; i-- {
		reverseDomain += "." + domainParts[i]
	}
	// reverseDomain += "."

	p.log.Debug(
		"dns zone found",
		zap.String("zone_id", p.ZoneID),
		zap.String("domain", domain),
		// zap.Any("record_prefix", reverseDomain),
	)

	// Get information about existing records
	recordUpdated := false
	recordSetRequest := &route53.ListResourceRecordSetsInput{}
	recordSetRequest.SetHostedZoneId(p.ZoneID)
	recordSetRequest.SetMaxItems("100")
	if err := recordSetRequest.Validate(); err != nil {
		return fmt.Errorf("list resource record sets request validation error: %s", err)
	}

	recordSetResponse, err := svc.ListResourceRecordSets(recordSetRequest)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				return fmt.Errorf("zone id %s not found: %s", p.ZoneID, aerr.Error())
			case route53.ErrCodeInvalidInput:
				return fmt.Errorf("invalid list resource record sets request in zone id %s: %s", p.ZoneID, aerr.Error())
			default:
				return fmt.Errorf("list resource record sets request failed: %s", aerr.Error())
			}
		}
		return fmt.Errorf("list resource record sets request failed: %s", err.Error())
	}

	var recordCurrentValue string
	for _, rrset := range recordSetResponse.ResourceRecordSets {
		if string(*rrset.Name) != fqdn {
			continue
		}
		if len(rrset.ResourceRecords) != 1 {
			continue
		}
		rr := rrset.ResourceRecords[0]
		if string(*rr.Value) != ip4 {
			recordCurrentValue = string(*rr.Value)
			continue
		}
		recordUpdated = true
	}

	if recordUpdated {
		p.log.Debug(
			"dns resource record set is up to date",
			zap.String("zone_id", p.ZoneID),
			zap.String("hostname", hostname),
			zap.String("fqdn", fqdn),
			zap.String("ip4", ip4),
		)
		return nil
	}

	p.log.Info(
		"dns resource record set is outdated",
		zap.String("zone_id", p.ZoneID),
		zap.String("hostname", hostname),
		zap.String("fqdn", fqdn),
		zap.String("outdated_ip4", recordCurrentValue),
		zap.String("ip4", ip4),
	)

	rr := &route53.ResourceRecord{}
	rr.SetValue(ip4)
	rrSet := &route53.ResourceRecordSet{}
	rrSet.SetName(fqdn)
	rrSet.SetType("A")
	rrSet.SetTTL(int64(r.TimeToLive))
	// rrSet.SetWeight(0)
	rrSet.SetResourceRecords([]*route53.ResourceRecord{rr})
	// rrSet.SetSetIdentifier("RR-A-" + strings.ToUpper(hostname))
	if err := rrSet.Validate(); err != nil {
		return fmt.Errorf("resource record set validation error: %s", err)
	}

	rrChange := &route53.Change{}
	rrChange.SetAction("UPSERT")
	rrChange.SetResourceRecordSet(rrSet)
	if err := rrChange.Validate(); err != nil {
		return fmt.Errorf("resource record change validation error: %s", err)
	}

	rrBatchChange := &route53.ChangeBatch{}
	rrBatchChange.SetChanges([]*route53.Change{rrChange})
	rrBatchChange.SetComment("dyndns updated on " + time.Now().String())
	if err := rrBatchChange.Validate(); err != nil {
		return fmt.Errorf("resource record change batch validation error: %s", err)
	}

	rrBatchChangeRequest := &route53.ChangeResourceRecordSetsInput{}
	rrBatchChangeRequest.SetHostedZoneId(p.ZoneID)
	rrBatchChangeRequest.SetChangeBatch(rrBatchChange)
	if err := rrBatchChangeRequest.Validate(); err != nil {
		return fmt.Errorf("resource record change batch validation error: %s", err)
	}
	rrBatchResponse, err := svc.ChangeResourceRecordSets(rrBatchChangeRequest)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				return fmt.Errorf("zone id %s not found: %s", p.ZoneID, aerr.Error())
			case route53.ErrCodeInvalidInput:
				return fmt.Errorf("invalid resource record change batch input in zone id %s: %s", p.ZoneID, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				return fmt.Errorf("invalid resource record change batch in zone id %s: %s", p.ZoneID, aerr.Error())
			}
		}
		return fmt.Errorf("resource record change batch request failed: %s", err.Error())
	}

	p.log.Info(
		"dns resource record updated",
		zap.String("zone_id", p.ZoneID),
		// zap.String("record_set", rrSet.String()),
		zap.String("status", *rrBatchResponse.ChangeInfo.Status),
		zap.String("hostname", hostname),
		zap.String("fqdn", fqdn),
		zap.String("ip4", ip4),
	)

	return nil
}
