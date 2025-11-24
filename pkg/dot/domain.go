package dot

import "github.com/yaklabco/dot/internal/domain"

// Domain entity re-exports

// Package represents a collection of configuration files to be managed.
type Package = domain.Package

// NodeType identifies the type of filesystem node.
type NodeType = domain.NodeType

// NodeType constants
const (
	NodeFile    = domain.NodeFile
	NodeDir     = domain.NodeDir
	NodeSymlink = domain.NodeSymlink
)

// Node represents a node in a filesystem tree.
type Node = domain.Node

// Plan represents a set of operations to execute.
type Plan = domain.Plan

// PlanMetadata contains statistics and diagnostic information about a plan.
type PlanMetadata = domain.PlanMetadata
