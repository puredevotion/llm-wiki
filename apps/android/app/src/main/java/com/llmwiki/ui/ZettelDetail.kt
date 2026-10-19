package com.llmwiki.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ZettelDetailScreen(
    zettelId: String,
    viewModel: ZettelViewModel,
    onBack: () -> Unit
) {
    val zettels by viewModel.zettels.collectAsState()
    val zettel = zettels.find { it.id == zettelId }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(zettel?.title ?: "Note") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        if (zettel != null) {
            Column(
                modifier = Modifier
                    .padding(padding)
                    .padding(16.dp)
                    .verticalScroll(rememberScrollState())
                    .fillMaxSize(),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Badge { Text(zettel.lifecycle) }
                Text(text = zettel.body, style = MaterialTheme.typography.bodyLarge)
                
                HorizontalDivider()
                
                Text(text = "Metadata", style = MaterialTheme.typography.titleSmall)
                Text(text = "Created by: ${zettel.created_by}", style = MaterialTheme.typography.labelMedium)
                Text(text = "Updated: ${zettel.updated_at}", style = MaterialTheme.typography.labelMedium)
            }
        } else {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = androidx.compose.ui.Alignment.Center) {
                Text("Zettel not found")
            }
        }
    }
}
