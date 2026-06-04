package api

func openAPIOperation(tag string, operationID string, summary string, description string, responseSchema string, requestSchema string) map[string]any {
	return openAPIOperationWithParameters(tag, operationID, summary, description, responseSchema, requestSchema, nil)
}

func openAPIOperationWithParameters(tag string, operationID string, summary string, description string, responseSchema string, requestSchema string, parameters []map[string]any) map[string]any {
	operation := map[string]any{
		"tags":        []string{tag},
		"operationId": operationID,
		"summary":     summary,
		"description": description,
		"responses": map[string]any{
			"400": openAPIErrorResponse("요청 값이 잘못됐습니다."),
			"404": openAPIErrorResponse("대상을 찾을 수 없습니다."),
			"409": openAPIErrorResponse("현재 상태에서 요청을 처리할 수 없습니다."),
			"500": openAPIErrorResponse("서버 내부 오류입니다."),
		},
	}
	responses := operation["responses"].(map[string]any)
	if responseSchema == "" {
		responses["200"] = map[string]any{"description": "요청이 성공했습니다."}
	} else {
		responses["200"] = openAPIJSONResponse("요청이 성공했습니다.", "#/components/schemas/"+responseSchema)
	}
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}
	if requestSchema != "" {
		operation["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/" + requestSchema},
				},
			},
		}
	}
	return operation
}

func openAPIJSONResponse(description string, schemaReference string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": schemaReference},
			},
		},
	}
}

func openAPIErrorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func openAPIObjectSchema(description string, properties map[string]any) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
}

func openAPIGenericObjectSchema(description string) map[string]any {
	return openAPIObjectSchema(description, map[string]any{
		"data": openAPIGenericObjectProperty("응답 payload입니다."),
	})
}

func openAPIGenericObjectProperty(description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          description,
		"additionalProperties": true,
	}
}

func openAPIArrayEnvelopeSchema(fieldName string, itemReference string, description string) map[string]any {
	return openAPIObjectSchema(description, map[string]any{
		fieldName: map[string]any{
			"type":  "array",
			"items": map[string]any{"$ref": itemReference},
		},
	})
}

func openAPIStringMapProperty(description string) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"additionalProperties": map[string]any{
			"type": "string",
		},
	}
}

func openAPIStringArrayProperty(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": description,
		"items":       map[string]any{"type": "string"},
	}
}

func openAPIBoolMapProperty(description string) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"additionalProperties": map[string]any{
			"type": "boolean",
		},
	}
}

func openAPIPathParameter(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func openAPIQueryParameter(name string, description string, required bool) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    required,
		"description": description,
		"schema":      map[string]any{"type": "string"},
	}
}

func openAPIIntegerQueryParameter(name string, description string, required bool) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    required,
		"description": description,
		"schema":      map[string]any{"type": "integer"},
	}
}

func openAPISensorQueryParameters(required bool) []map[string]any {
	return []map[string]any{
		openAPIQueryParameter("missionId", "mission ID 또는 missionCode입니다.", required),
		openAPIQueryParameter("robotCode", "robot code입니다.", required),
	}
}

func openAPIStringProperty(description string, example string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
		"example":     example,
	}
}

func openAPINumberProperty(description string, example float64) map[string]any {
	return map[string]any{
		"type":        "number",
		"description": description,
		"example":     example,
	}
}

func openAPIDateTimeProperty(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"format":      "date-time",
		"description": description,
		"example":     "2026-06-01T11:30:00Z",
	}
}

func openAPINullableDateTimeProperty(description string) map[string]any {
	property := openAPIDateTimeProperty(description)
	property["nullable"] = true
	return property
}
