package com.llmwiki.di

import com.llmwiki.data.*
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object RepositoryModule {

    @Provides
    @Singleton
    fun provideSyncRepository(
        api: LLMWikiApi,
        zettelDao: ZettelDao,
        eventDao: EventDao,
        syncDao: SyncDao
    ): SyncRepository {
        return SyncRepository(api, zettelDao, eventDao, syncDao)
    }
}
