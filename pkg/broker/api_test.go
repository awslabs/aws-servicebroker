package broker

import (
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCatalog(t *testing.T) {
	assertor := assert.New(t)

	opts := Options{
		TableName:          "testtable",
		S3Bucket:           "abucket",
		S3Region:           "us-east-1",
		S3Key:              "tempates/test",
		Region:             "us-east-1",
		BrokerID:           "awsservicebroker",
		PrescribeOverrides: false,
	}
	bl, _ := NewAWSBroker(opts, mockGetAwsSession, mockClients, mockGetAccountId, mockUpdateCatalog, mockPollUpdate)
	bl.listingcache.Set("__LISTINGS__", []ServiceNeedsUpdate{{Name: "test", Update: false}})

	expected := &broker.CatalogResponse{
		CatalogResponse: osb.CatalogResponse{},
	}
	actual, err := bl.GetCatalog(&broker.RequestContext{})
	assertor.Equal(nil, err, "err should be nil")
	assertor.Equal(expected, actual, "should return empty catalog")
}
