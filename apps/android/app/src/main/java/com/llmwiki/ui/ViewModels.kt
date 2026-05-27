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
