package com.llmwiki.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.llmwiki.domain.Actor
import com.llmwiki.domain.Team

@OptIn(ExperimentalMaterial3Api::class)
fun IdentityScreen(
    viewModel: IdentityViewModel
) {
    val actors by viewModel.actors.collectAsState()
    val teams by viewModel.teams.collectAsState()
    val organizations by viewModel.organizations.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Who") })
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            item {
                Text(text = "Organizations", style = MaterialTheme.typography.titleLarge)
            }
            items(organizations) { org ->
                IdentityItem(title = org.name, subtitle = "Organization", icon = Icons.Default.Star)
            }

            item {
                Spacer(modifier = Modifier.height(16.dp))
                Text(text = "Teams", style = MaterialTheme.typography.titleLarge)
            }
...

            items(teams) { team ->
                IdentityItem(title = team.name, subtitle = "Team", icon = Icons.Default.Star)
            }
            
            item {
                Spacer(modifier = Modifier.height(16.dp))
                Text(text = "People", style = MaterialTheme.typography.titleLarge)
            }
            items(actors) { actor ->
                IdentityItem(title = actor.display_name, subtitle = actor.kind, icon = Icons.Default.Person)
            }
        }
    }
}

@Composable
fun IdentityItem(title: String, subtitle: String, icon: ImageVector) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(icon, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
            Spacer(modifier = Modifier.width(16.dp))
            Column {
                Text(text = title, style = MaterialTheme.typography.titleMedium)
                Text(text = subtitle, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}
