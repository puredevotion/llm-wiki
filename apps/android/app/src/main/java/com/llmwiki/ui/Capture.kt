package com.llmwiki.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CaptureScreen(
    onBack: () -> Unit,
    onSave: (title: String, body: String, kind: String) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var content by remember { mutableStateOf("") }
    var selectedKind by remember { mutableStateOf("zettel") }
    val kinds = listOf("zettel", "chat_excerpt", "document", "meeting_note")

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Capture Seed") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { onSave(title, content, selectedKind) }) {
                        Icon(Icons.Default.Check, contentDescription = "Save")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .padding(16.dp)
                .verticalScroll(rememberScrollState())
                .fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Text("What are you seeding today?", style = MaterialTheme.typography.titleMedium)
            
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                kinds.forEach { kind ->
                    FilterChip(
                        selected = selectedKind == kind,
                        onClick = { selectedKind = kind },
                        label = { Text(kind.replace("_", " ").capitalize()) }
                    )
                }
            }

            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Title") },
                modifier = Modifier.fillMaxWidth()
            )

            OutlinedTextField(
                value = content,
                onValueChange = { content = it },
                label = { Text("Content / Excerpt") },
                modifier = Modifier
                    .fillMaxWidth()
                    .heightIn(min = 200.dp),
                placeholder = { Text("Paste your chat log or type your note here...") }
            )

            Button(
                onClick = { onSave(title, content, selectedKind) },
                modifier = Modifier.fillMaxWidth(),
                enabled = title.isNotBlank() && content.isNotBlank()
            ) {
                Text("Seed Knowledge")
            }
        }
    }
}

fun String.capitalize() = replaceFirstChar { if (it.isLowerCase()) it.titlecase() else it.toString() }
