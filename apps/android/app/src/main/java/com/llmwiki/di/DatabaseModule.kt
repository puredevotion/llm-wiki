package com.llmwiki.di

import android.content.Context
import androidx.room.Room
import com.llmwiki.data.*
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): LLMWikiDatabase {
        return Room.databaseBuilder(
            context,
            LLMWikiDatabase::class.java,
            "llm_wiki.db"
        ).build()
    }

    @Provides
    fun provideZettelDao(db: LLMWikiDatabase): ZettelDao = db.zettelDao()

    @Provides
    fun provideEventDao(db: LLMWikiDatabase): EventDao = db.eventDao()

    @Provides
    fun provideSyncDao(db: LLMWikiDatabase): SyncDao = db.syncDao()

    @Provides
    fun provideActorDao(db: LLMWikiDatabase): ActorDao = db.actorDao()

    @Provides
    fun provideTeamDao(db: LLMWikiDatabase): TeamDao = db.teamDao()

    @Provides
    fun provideOrganizationDao(db: LLMWikiDatabase): OrganizationDao = db.organizationDao()
}
