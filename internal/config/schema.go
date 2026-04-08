package config

import (
	pluginv1 "github.com/ContinuumApp/continuum-plugin-sdk/pkg/pluginproto/continuum/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type ConfigSchema = pluginv1.ConfigSchema

func GlobalConfigSchema() []*ConfigSchema {
	return []*ConfigSchema{
		objectSchema("general", "General", "General Dispatcharr plugin settings.", `{"type":"object","properties":{"source_mode":{"type":"string","enum":["xtream","m3u_xmltv"]},"live_tv_enabled":{"type":"boolean"}},"required":["source_mode"],"additionalProperties":false}`, true, []*pluginv1.AdminFormField{
			{Key: "source_mode", Label: "Source Mode", Description: "Choose the Dispatcharr source mode.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_SELECT, Required: true, DefaultValue: structpb.NewStringValue(string(SourceModeXtream)), Options: []*pluginv1.AdminFormOption{{Value: string(SourceModeXtream), Label: "Xtream"}, {Value: string(SourceModeM3UXMLTV), Label: "M3U/XMLTV"}}},
			{Key: "live_tv_enabled", Label: "Enable Live TV", Description: "Expose the Live TV source to Continuum.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_SWITCH, DefaultValue: structpb.NewBoolValue(true)},
		}, "Save general settings"),
		objectSchema("xtream", "Xtream", "Xtream connection settings for Dispatcharr.", `{"type":"object","properties":{"base_url":{"type":"string","format":"uri"},"username":{"type":"string","minLength":1},"password":{"type":"string","minLength":1,"writeOnly":true}},"required":["base_url","username","password"],"additionalProperties":false}`, true, []*pluginv1.AdminFormField{
			{Key: "base_url", Label: "Xtream Base URL", Description: "Dispatcharr Xtream endpoint base URL.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_TEXT, Placeholder: "https://dispatcharr.example.com", Required: true},
			{Key: "username", Label: "Xtream Username", Description: "Xtream username for Dispatcharr.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_TEXT, Required: true},
			{Key: "password", Label: "Xtream Password", Description: "Xtream password for Dispatcharr.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_PASSWORD, Required: true, Secret: true},
		}, "Save Xtream settings"),
		objectSchema("m3u_xmltv", "M3U/XMLTV", "Fallback playlist and XMLTV settings.", `{"type":"object","properties":{"m3u_url":{"type":"string","format":"uri"},"epg_xml_url":{"type":"string","format":"uri"}},"required":["m3u_url","epg_xml_url"],"additionalProperties":false}`, true, []*pluginv1.AdminFormField{
			{Key: "m3u_url", Label: "M3U URL", Description: "Playlist URL for fallback mode.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_TEXT, Placeholder: "https://dispatcharr.example.com/playlist.m3u", Required: true},
			{Key: "epg_xml_url", Label: "EPG XML URL", Description: "XMLTV URL for fallback mode.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_TEXT, Placeholder: "https://dispatcharr.example.com/guide.xml", Required: true},
		}, "Save M3U/XMLTV settings"),
		objectSchema("vod", "VOD", "Coming soon. VOD is not exposed in v1.", `{"type":"object","properties":{},"additionalProperties":false}`, false, []*pluginv1.AdminFormField{{Key: "status", Label: "VOD", Description: "Coming soon.", Control: pluginv1.AdminFormControl_ADMIN_FORM_CONTROL_TEXT, DefaultValue: structpb.NewStringValue("Coming soon")}}, "Save VOD settings"),
	}
}

func objectSchema(key, title, description, jsonSchema string, required bool, fields []*pluginv1.AdminFormField, submitLabel string) *ConfigSchema {
	return &pluginv1.ConfigSchema{
		Key:         key,
		Title:       title,
		Description: description,
		JsonSchema:  jsonSchema,
		Required:    required,
		AdminForm:   &pluginv1.AdminFormDescriptor{Fields: fields, SubmitLabel: submitLabel},
	}
}
