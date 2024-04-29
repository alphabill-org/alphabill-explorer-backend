// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/blocks": {
            "get": {
                "description": "Get blocks, given a start block number and limit.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Energy Type"
                ],
                "summary": "Get blocks, given a start block number and limit.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "optionally specify the start block number",
                        "name": "startBlock",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "optionally specify the number of blocks to return, defaults to 10",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/api.BlockInfo"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.BlockInfo": {
            "type": "object",
            "properties": {
                "header": {
                    "$ref": "#/definitions/types.Header"
                },
                "txHashes": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "unicityCertificate": {
                    "$ref": "#/definitions/types.UnicityCertificate"
                }
            }
        },
        "imt.PathItem": {
            "type": "object",
            "properties": {
                "hash": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "key": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "types.Header": {
            "type": "object",
            "properties": {
                "previousBlockHash": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "proposerID": {
                    "type": "string"
                },
                "shardID": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "systemID": {
                    "type": "integer"
                }
            }
        },
        "types.InputRecord": {
            "type": "object",
            "properties": {
                "block_hash": {
                    "description": "hash of the block",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "hash": {
                    "description": "state hash to be certified",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "previous_hash": {
                    "description": "previously certified state hash",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "round_number": {
                    "description": "transaction system's round number",
                    "type": "integer"
                },
                "sum_of_earned_fees": {
                    "description": "sum of the actual fees over all transaction records in the block",
                    "type": "integer"
                },
                "summary_value": {
                    "description": "summary value to certified",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "types.SignatureMap": {
            "type": "object",
            "additionalProperties": {
                "type": "array",
                "items": {
                    "type": "integer"
                }
            }
        },
        "types.UnicityCertificate": {
            "type": "object",
            "properties": {
                "input_record": {
                    "$ref": "#/definitions/types.InputRecord"
                },
                "unicity_seal": {
                    "$ref": "#/definitions/types.UnicitySeal"
                },
                "unicity_tree_certificate": {
                    "$ref": "#/definitions/types.UnicityTreeCertificate"
                }
            }
        },
        "types.UnicitySeal": {
            "type": "object",
            "properties": {
                "hash": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "previous_hash": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "root_chain_round_number": {
                    "type": "integer"
                },
                "signatures": {
                    "$ref": "#/definitions/types.SignatureMap"
                },
                "timestamp": {
                    "type": "integer"
                }
            }
        },
        "types.UnicityTreeCertificate": {
            "type": "object",
            "properties": {
                "sibling_hashes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/imt.PathItem"
                    }
                },
                "system_description_hash": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "system_identifier": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Alphabill Blockchain Explorer API",
	Description:      "API to query blocks and transactions of Alphabill's Money Partition",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
