package com.llmwiki.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.llmwiki.data.SyncRepository
import com.llmwiki.domain.Zettel
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ZettelViewModel @Inject constructor(
    private val repository: SyncRepository
) : ViewModel() {

    val zettels: StateFlow<List<Zettel>> = repository.getZettels()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    fun refresh() {
        viewModelScope.launch {
            try {
                repository.sync()
            } catch (e: Exception) {
                // Handle error
            }
        }
    }
}

@HiltViewModel
class TimelineViewModel @Inject constructor(
    private val repository: SyncRepository
) : ViewModel() {

    val events: StateFlow<List<TimelineEvent>> = repository.getEvents()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())
}

@HiltViewModel
class IdentityViewModel @Inject constructor(
    private val repository: SyncRepository
) : ViewModel() {

    val actors: StateFlow<List<com.llmwiki.domain.Actor>> = repository.getActors()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val teams: StateFlow<List<com.llmwiki.domain.Team>> = repository.getTeams()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val organizations: StateFlow<List<com.llmwiki.domain.Organization>> = repository.getOrganizations()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    fun capture(title: String, body: String, kind: String) {
        viewModelScope.launch {
            repository.capture(title, body, kind)
        }
    }
}

@HiltViewModel
class GraphViewModel @Inject constructor(
    private val api: com.llmwiki.data.LLMWikiApi
) : ViewModel() {

    private val _graphData = kotlinx.coroutines.flow.MutableStateFlow<com.llmwiki.data.GraphDataResponse?>(null)
    val graphData: StateFlow<com.llmwiki.data.GraphDataResponse?> = _graphData

    init {
        fetchGraph()
    }

    fun fetchGraph() {
        viewModelScope.launch {
            try {
                _graphData.value = api.getGraph()
            } catch (e: Exception) {
                // Handle error
            }
        }
    }
}
