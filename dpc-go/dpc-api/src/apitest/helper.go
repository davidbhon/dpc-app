package apitest

import (
	"encoding/json"
	"fmt"

	"github.com/CMSgov/dpc/api/model"
	"github.com/bxcodec/faker/v3"
	luhn "github.com/joeljunstrom/go-luhn"
)

// Orgjson is a organization json string for testing purposes
const Orgjson = `{
  "resourceType": "Organization",
  "identifier": [
    {
      "use": "official",
      "system": "urn:oid:2.16.528.1",
      "value": "91654"
    },
    {
      "use": "usual",
      "system": "urn:oid:2.16.840.1.113883.2.4.6.1",
      "value": "17-0112278"
    }
  ],
  "type": [
    {
      "coding": [
        {
          "system": "urn:oid:2.16.840.1.113883.2.4.15.1060",
          "code": "V6",
          "display": "University Medical Hospital"
        },
        {
          "system": "http://terminology.hl7.org/CodeSystem/organization-type",
          "code": "prov",
          "display": "Healthcare Provider"
        }
      ]
    }
  ],
  "name": "Burgers University Medical Center",
  "telecom": [
    {
      "system": "phone",
      "value": "022-655 2300",
      "use": "work"
    }
  ],
  "address": [
    {
      "use": "work",
      "line": [
        "Galapagosweg 91"
      ],
      "city": "Den Burg",
      "postalCode": "9105 PZ",
      "country": "NLD"
    },
    {
      "use": "work",
      "line": [
        "PO Box 2311"
      ],
      "city": "Den Burg",
      "postalCode": "9100 AA",
      "country": "NLD"
    }
  ],
  "contact": [
    {
      "purpose": {
        "coding": [
          {
            "system": "http://terminology.hl7.org/CodeSystem/contactentity-type",
            "code": "PRESS"
          }
        ]
      },
      "telecom": [
        {
          "system": "phone",
          "value": "022-655 2334"
        }
      ]
    },
    {
      "purpose": {
        "coding": [
          {
            "system": "http://terminology.hl7.org/CodeSystem/contactentity-type",
            "code": "PATINF"
          }
        ]
      },
      "telecom": [
        {
          "system": "phone",
          "value": "022-655 2335"
        }
      ]
    }
  ]
}`

// Groupjson is a group json string for testing purposes
const Groupjson = `
{
  "resourceType": "Group",
  "id": "fullexample",
  "meta": {
    "versionId": "1",
    "lastUpdated": "2019-06-06T03:04:12.348-04:00"
  },
  "extension": [
    {
      "url": "http://hl7.org/fhir/us/davinci-atr/StructureDefinition/ext-contractValidityPeriod",
      "valuePeriod": {
        "start": "2020-07-25",
        "end": "2021-06-24"
      }
    }
  ],
  "identifier": [
    {
      "use": "official",
      "type": {
        "coding": [
          {
            "system": "http://terminology.hl7.org/CodeSystem/v2-0203",
            "code": "NPI",
            "display": "National Provider Identifier"
          }
        ]
      },
      "system": "https://sitenv.org",
      "value": "1316206220"
    },
    {
      "use": "official",
      "type": {
        "coding": [
          {
            "system": "http://terminology.hl7.org/CodeSystem/v2-0203",
            "code": "TAX",
            "display": "Tax ID Number"
          }
        ]
      },
      "system": "https://sitenv.org",
      "value": "789456231"
    }
  ],
  "active": true,
  "type": "person",
  "actual": true,
  "name": "Test Group 3",
  "managingEntity": {
    "reference": "Organization/1",
    "display": "Healthcare related organization"
  },
  "member": [
    {
      "extension": [
        {
          "url": "http://hl7.org/fhir/us/davinci-atr/StructureDefinition/ext-changeType",
          "valueCode": "add"
        },
        {
          "url": "http://hl7.org/fhir/us/davinci-atr/StructureDefinition/ext-coverageReference",
          "valueReference": {
            "reference": "Coverage/1"
          }
        },
        {
          "url": "http://hl7.org/fhir/us/davinci-atr/StructureDefinition/ext-attributedProvider",
          "valueReference": {
            "type": "Practitioner",
            "identifier": {
                "system": "http://hl7.org/fhir/sid/us-npi",
                "value": "9941339108"
            }
          }
        }
      ],
      "entity": {
        "type": "Patient",
        "identifier": {
            "value": "2SW4N00AA00",
            "system": "http://hl7.org/fhir/sid/us-mbi"
        }
      },
      "period": {
        "start": "2014-10-08",
        "end": "2020-10-08"
      },
      "inactive": false
    }
  ]
}`

// FilteredGroupjson is a group json string for testing purposes
const FilteredGroupjson = `
{
  "resourceType": "Group",
  "type": "person",
  "actual": true,
  "name": "Test Group 3",
  "managingEntity": {
    "reference": "Organization/1",
    "display": "Healthcare related organization"
  },
  "member": [
    {
      "extension": [
        {
          "url": "http://hl7.org/fhir/us/davinci-atr/StructureDefinition/ext-attributedProvider",
          "valueReference": {
            "type": "Practitioner",
            "identifier": {
                "system": "http://hl7.org/fhir/sid/us-npi",
                "value": "9941339108"
            }
          }
        }
      ],
      "entity": {
        "type": "Patient",
        "identifier": {
            "value": "2SW4N00AA00",
            "system": "http://hl7.org/fhir/sid/us-mbi"
        }
      },
      "period": {
        "start": "2014-10-08",
        "end": "2020-10-08"
      },
      "inactive": false
    }
  ]
}`

// JobJSON is a Job json string for testing purposes
const JobJSON = `{
  "id": "test-export-job"
}`

// ImplJSON is an Implementer json string for testing purposes
const ImplJSON = `{
    "name": "foo"
}`

// GetImplOrgJSON is an Implementer Organization JSON string for testing purposes
const GetImplOrgJSON = `[
    {
        "org_id": "5c3934c0-6c7e-42e1-b50f-2086b801680b",
        "org_name": "Huge Xylophone Healthcare",
        "status": "Active",
        "npi": "0123456789",
        "ssas_system_id": ""
    }
]`

// GetBatchAndFilesJSON is an status example from attribution
const GetBatchAndFilesJSON = `[
  {
    "batch": {
      "totalPatients": 32,
      "patientsProcessed": 32,
      "patientIndex": 31,
      "status": "COMPLETED",
      "transactionTime": "2021-08-02T00:00:00Z",
      "submitTime": "2021-08-02T00:00:00Z",
      "completeTime": "2021-08-02T00:00:00Z",
      "requestURL": "http://pfSbNLv.info/"
    },
    "files": [
      {
        "resourceType": "Patient",
        "batchID": "f9185824-c835-421d-81f9-ec2b1ee609af",
        "sequence": 0,
        "fileName": "testFileName",
        "count": 1,
        "checksum": "ad09ae2eee0a5111508b072cb8c3eaca49f342df82b7c456bcd04df7612283e77f04f0d55e89a5a235f6636a3a9180169b890f6a3078e200fd9a1ca1574885767ffa30ddf94bc374464cf8c6f1da72c8",
        "fileLength": 1234
      }
    ]
  }
]`

// ImplOrgJSON creates an Implementer/Org JSON string for testing purposes
func ImplOrgJSON() string {
	body := struct {
		Npi string `json:"npi"`
	}{
		Npi: GenerateNPI(),
	}
	jsonData, _ := json.Marshal(body)
	return string(jsonData)
}

// AttributionOrgResponse provides a sample organization response that mimics what attribution service returns for testing purposes
func AttributionOrgResponse() []byte {
	return AttributionToFHIRResponse(Orgjson)
}

// AttributionToFHIRResponse provides a sample response that mimics what attribution service returns for testing purposes
func AttributionToFHIRResponse(fhir string) []byte {
	r := model.Resource{}
	err := faker.FakeData(&r)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}

	var v map[string]interface{}
	_ = json.Unmarshal([]byte(fhir), &v)
	r.Info = v
	// The value <<PRESENCE>> required for jsonassert checks
	r.ID = "<<PRESENCE>>"
	b, _ := json.Marshal(r)
	return b
}

// AttributionResponse provides a sample response that mimics what attribution service returns for testing purposes
func AttributionResponse(data string) []byte {
	var v map[string]interface{}
	_ = json.Unmarshal([]byte(data), &v)
	b, _ := json.Marshal(v)
	return b
}

// MalformedOrg provides a convenience method to get a non valid fhir resource, in this case an org
func MalformedOrg() []byte {
	var org map[string]interface{}
	_ = json.Unmarshal([]byte(Orgjson), &org)
	org["resourceType"] = "trash"
	b, _ := json.Marshal(org)
	return b
}

// GenerateNPI creates a placeholder NPI number for testing purposes
func GenerateNPI() string {
	luhnWithPrefix := luhn.GenerateWithPrefix(15, "808403")
	return luhnWithPrefix[len(luhnWithPrefix)-10:]
}

// ToBytes converts an interface to bytes
func ToBytes(a interface{}) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		fmt.Println(err)
	}
	return b
}
