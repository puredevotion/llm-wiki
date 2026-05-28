package com.llmwiki.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.llmwiki.domain.TimelineEvent

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimelineScreen(
    viewModel: TimelineViewModel
) {
    val events by viewModel.events.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Timeline") })
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            items(events) { event ->
                EventItem(event = event)
            }
        }
    }
}

@Composable
fun EventItem(event: TimelineEvent) {
    Row(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.width(64.dp)) {
            Text(
                text = event.occurred_at?.take(10) ?: "Log",
                style = MaterialTheme.typography.labelSmall,
                fontWeight = FontWeight.Bold
            )
        }
        Column(modifier = Modifier.weight(1f)) {
            Text(text = event.title, style = MaterialTheme.typography.titleMedium)
            Text(
                text = event.body,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Badge(modifier = Modifier.padding(top = 4.dp)) {
                Text(event.kind)
            }
        }
    }
}
