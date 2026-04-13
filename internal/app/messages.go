package app

// Message types have moved to github.com/seanhalberthal/seeql/internal/msg.
// This file re-exports them for convenience within the app package.

import appmsg "github.com/seanhalberthal/seeql/internal/msg"

// Re-export types used within app package.
type (
	Pane              = appmsg.Pane
	ConnectMsg        = appmsg.ConnectMsg
	ConnectErrMsg     = appmsg.ConnectErrMsg
	DisconnectMsg     = appmsg.DisconnectMsg
	SchemaLoadedMsg   = appmsg.SchemaLoadedMsg
	SchemaErrMsg      = appmsg.SchemaErrMsg
	ExecuteQueryMsg   = appmsg.ExecuteQueryMsg
	QueryStartedMsg   = appmsg.QueryStartedMsg
	QueryResultMsg    = appmsg.QueryResultMsg
	QueryErrMsg       = appmsg.QueryErrMsg
	QueryStreamingMsg = appmsg.QueryStreamingMsg
	NewTabMsg         = appmsg.NewTabMsg
	CloseTabMsg       = appmsg.CloseTabMsg
	SwitchTabMsg      = appmsg.SwitchTabMsg
	StatusMsg         = appmsg.StatusMsg
	InsertTextMsg     = appmsg.InsertTextMsg
	ExportCompleteMsg = appmsg.ExportCompleteMsg
	ExportErrMsg      = appmsg.ExportErrMsg
)

// Re-export constants.
const (
	PaneSidebar = appmsg.PaneSidebar
	PaneEditor  = appmsg.PaneEditor
	PaneResults = appmsg.PaneResults
)
