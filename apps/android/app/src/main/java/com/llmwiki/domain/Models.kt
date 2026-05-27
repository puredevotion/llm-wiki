package com.llmwiki.domain

data class Zettel(
    val id: String,
    val title: String,
    val body: String,
    val lifecycle: String,
    val status: String,
    val created_by: String,
    val created_at: String,
    val updated_at: String
)

data class TimelineEvent(
    val id: String,
    val kind: String,
    val title: String,
    val body: String,
    val occurred_at: String?,
    val starts_at: String?,
    val ends_at: String?,
    val recorded_at: String,
    val created_by: String,
    val metadata: Map<String, Any>
)

data class SyncOperation(
    val id: String,
    val actor_id: String,
    val device_id: String,
    val entity_kind: String,
    val entity_id: String,
    val operation_type: String,
    val payload: Any, // Raw payload
    val base_version: Int,
    val status: String,
    val created_at: String
)

data class Actor(
    val id: String,
    val kind: String,
    val display_name: String,
    val created_at: String,
    val metadata: Map<String, Any>
)

data class Team(
    val id: String,
    val org_id: String?,
    val name: String,
    val created_at: String,
    val metadata: Map<String, Any>
)
