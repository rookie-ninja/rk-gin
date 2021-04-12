// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/rk/apis": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "List APIs",
                "operationId": "5",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/rkgin.APIsResponse"
                            }
                        }
                    }
                }
            }
        },
        "/v1/rk/config": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "List Configs",
                "operationId": "4",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkgin.ConfigResponse"
                        }
                    }
                }
            }
        },
        "/v1/rk/gc": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Trigger GC",
                "operationId": "2",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkgin.GCResponse"
                        }
                    }
                }
            }
        },
        "/v1/rk/healthy": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Check healthy status",
                "operationId": "1",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkgin.HealthyResponse"
                        }
                    }
                }
            }
        },
        "/v1/rk/info": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Service Info",
                "operationId": "3",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkinfo.BasicInfo"
                        }
                    }
                }
            }
        },
        "/v1/rk/req": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Request Stat",
                "operationId": "7",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkmetrics.ReqMetricsRK"
                        }
                    }
                }
            }
        },
        "/v1/rk/sys": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "System Stat",
                "operationId": "6",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/rkgin.SysResponse"
                        }
                    }
                }
            }
        },
        "/v1/rk/tv": {
            "get": {
                "produces": [
                    "text/html"
                ],
                "summary": "HTML page",
                "operationId": "8",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "rkgin.APIsResponse": {
            "type": "object",
            "properties": {
                "method": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                },
                "sw_url": {
                    "type": "string"
                }
            }
        },
        "rkgin.ConfigResponse": {
            "type": "object",
            "properties": {
                "viper": {
                    "type": "string"
                }
            }
        },
        "rkgin.GCResponse": {
            "type": "object",
            "properties": {
                "mem_stat_after_gc": {
                    "$ref": "#/definitions/rkinfo.MemStats"
                },
                "mem_stat_before_gc": {
                    "$ref": "#/definitions/rkinfo.MemStats"
                }
            }
        },
        "rkgin.HealthyResponse": {
            "type": "object",
            "properties": {
                "healthy": {
                    "type": "boolean"
                }
            }
        },
        "rkgin.SysResponse": {
            "type": "object",
            "properties": {
                "cpu_usage_percentage": {
                    "type": "number"
                },
                "mem_usage_mb": {
                    "type": "integer"
                },
                "mem_usage_percentage": {
                    "type": "number"
                },
                "sys_up_time": {
                    "type": "string"
                }
            }
        },
        "rkinfo.BasicInfo": {
            "type": "object",
            "properties": {
                "application_name": {
                    "type": "string"
                },
                "az": {
                    "type": "string"
                },
                "domain": {
                    "type": "string"
                },
                "gid": {
                    "type": "string"
                },
                "realm": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                },
                "start_time": {
                    "type": "string"
                },
                "uid": {
                    "type": "string"
                },
                "up_time_sec": {
                    "type": "integer"
                },
                "up_time_str": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "rkinfo.MemStats": {
            "type": "object",
            "properties": {
                "force_gc_count": {
                    "type": "integer"
                },
                "gc_count_total": {
                    "type": "integer"
                },
                "last_gc_timestamp": {
                    "type": "string"
                },
                "mem_alloc_byte": {
                    "type": "integer"
                },
                "mem_usage_percentage": {
                    "type": "number"
                },
                "sys_alloc_byte": {
                    "type": "integer"
                }
            }
        },
        "rkmetrics.ReqMetricsRK": {
            "type": "object",
            "properties": {
                "count": {
                    "type": "integer"
                },
                "elapsed_nano_p50": {
                    "type": "number"
                },
                "elapsed_nano_p90": {
                    "type": "number"
                },
                "elapsed_nano_p99": {
                    "type": "number"
                },
                "elapsed_nano_p999": {
                    "type": "number"
                },
                "path": {
                    "type": "string"
                },
                "res_code": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/rkmetrics.ResCodeRK"
                    }
                }
            }
        },
        "rkmetrics.ResCodeRK": {
            "type": "object",
            "properties": {
                "count": {
                    "type": "integer"
                },
                "res_code": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "RK Swagger Example",
	Description: "This is a common service with rk-gin.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
