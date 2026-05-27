package com.llmwiki.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RelationsScreen(
    viewModel: GraphViewModel
) {
    val data by viewModel.graphData.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Relations") },
                actions = {
                    IconButton(onClick = { viewModel.fetchGraph() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        }
    ) { padding ->
        if (data == null) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = androidx.compose.ui.Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                item {
                    Text(text = "Entities & Links", style = MaterialTheme.typography.titleLarge)
                    Spacer(modifier = Modifier.height(16.dp))
                }
                
                items(data!!.links) { link ->
                    val source = data!!.nodes.find { it.id == link.source }
                    val target = data!!.nodes.find { it.id == link.target }
                    
                    RelationCard(
                        sourceName = source?.name ?: "Unknown",
                        targetName = target?.name ?: "Unknown",
                        relType = link.type
                    )
                }
            }
        }
    }
}

@Composable
fun RelationCard(sourceName: String, targetName: String, relType: String) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = androidx.compose.ui.Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(text = sourceName, style = MaterialTheme.typography.titleSmall)
                Text(text = "Source", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            
            Column(horizontalAlignment = androidx.compose.ui.Alignment.CenterHorizontally, modifier = Modifier.padding(horizontal = 8.dp)) {
                Icon(Icons.Default.Share, contentDescription = null, modifier = Modifier.size(16.dp))
                Text(text = relType, style = MaterialTheme.typography.labelExtraSmall())
            }
            
            Column(modifier = Modifier.weight(1f), horizontalAlignment = androidx.compose.ui.Alignment.End) {
                Text(text = targetName, style = MaterialTheme.typography.titleSmall)
                Text(text = "Target", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}

@Composable
fun MaterialTheme.typography.labelExtraSmall() = labelSmall.copy(fontSize = androidx.compose.ui.unit.TextUnit.Unspecified) 
// Simplified helper for very small text
