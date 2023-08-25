package models

// GroupsCache groupID -> userID -> groupMember
type GroupsCache map[int64]SingleGroupCache
type SingleGroupCache map[int64]*GroupMember
