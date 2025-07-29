//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2025 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package discovery

// EventType is the type of event
//
//go:generate stringer -type EventType -linecomment=true
type EventType int

const (
	// UnknownEventType is the unknown event type
	UnknownEventType EventType = iota
	// Create is the create event type
	Create
	// Delete is the delete event type
	Delete
	// Update is the update event type
	Update
)
