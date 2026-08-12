// Package docs Swagger Specification for Herald API
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": ["http"],
    "swagger": "2.0",
    "info": {
        "description": "Herald Multi-Tenant Notification & Webhook Delivery API for Forahia Solutions",
        "title": "Herald API",
        "version": "1.0.0"
    },
    "host": "localhost:8080",
    "basePath": "/",
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Format: Bearer hrld_live_testkey123"
        }
    },
    "paths": {
        "/health": {
            "get": {
                "summary": "Health Check Endpoint",
                "responses": {
                    "200": { "description": "Service is healthy" }
                }
            }
        },
        "/api/v1/notifications": {
            "post": {
                "summary": "Create and queue a notification",
                "security": [{"BearerAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "required": ["channel", "recipient", "body"],
                            "properties": {
                                "channel": { "type": "string", "example": "email" },
                                "recipient": { "type": "string", "example": "patient@auramed.cc" },
                                "subject": { "type": "string", "example": "Lab Results Ready" },
                                "body": { "type": "string", "example": "Your medical report is available in AuraMed." },
                                "priority": { "type": "string", "example": "high" }
                            }
                        }
                    }
                ],
                "responses": {
                    "201": { "description": "Notification queued successfully" },
                    "401": { "description": "Unauthorized" }
                }
            },
            "get": {
                "summary": "List notifications for tenant",
                "security": [{"BearerAuth": []}],
                "responses": {
                    "200": { "description": "List of notifications" }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Herald API",
	Description:      "Multi-Tenant Notification & Webhook Delivery API.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
