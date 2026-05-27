package com.llmwiki.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MainApp()
        }
    }
}

@Composable
fun MainApp() {
    val navController = rememberNavController()
    val items = listOf(
        Screen.Garden,
        Screen.Timeline,
        Screen.Relations,
        Screen.Identity
    )

    MaterialTheme {
        Scaffold(
            bottomBar = {
                val navBackStackEntry by navController.currentBackStackEntryAsState()
                val currentDestination = navBackStackEntry?.destination
                val showBottomBar = currentDestination?.route in items.map { it.route }
                
                if (showBottomBar) {
                    NavigationBar {
                        items.forEach { screen ->
                            NavigationBarItem(
                                icon = { Icon(screen.icon, contentDescription = null) },
                                label = { Text(screen.label) },
                                selected = currentDestination?.hierarchy?.any { it.route == screen.route } == true,
                                onClick = {
                                    navController.navigate(screen.route) {
                                        popUpTo(navController.graph.findStartDestination().id) {
                                            saveState = true
                                        }
                                        launchSingleTop = true
                                        restoreState = true
                                    }
                                }
                            )
                        }
                    }
                }
            }
        ) { innerPadding ->
            NavHost(
                navController = navController,
                startDestination = Screen.Garden.route,
                modifier = Modifier.padding(innerPadding)
            ) {
                composable(Screen.Garden.route) {
                    val viewModel: ZettelViewModel = hiltViewModel()
                    ZettelGardenScreen(
                        viewModel = viewModel,
                        onZettelClick = { id -> navController.navigate("zettel/$id") },
                        onAddClick = { navController.navigate("capture") }
                    )
                }
                composable(Screen.Timeline.route) {
                    val viewModel: TimelineViewModel = hiltViewModel()
                    TimelineScreen(viewModel = viewModel)
                }
                composable(Screen.Identity.route) {
                    val viewModel: IdentityViewModel = hiltViewModel()
                    IdentityScreen(viewModel = viewModel)
                }
                composable(Screen.Relations.route) {
                    val viewModel: GraphViewModel = hiltViewModel()
                    RelationsScreen(viewModel = viewModel)
                }
                composable("capture") {
                    CaptureScreen(
                        onBack = { navController.popBackStack() },
                        onSave = { title, body, kind ->
                            // TODO: Implement actual save logic in ViewModel
                            navController.popBackStack()
                        }
                    )
                }
                composable("zettel/{zettelId}") { backStackEntry ->
                    val zettelId = backStackEntry.arguments?.getString("zettelId") ?: ""
                    val viewModel: ZettelViewModel = hiltViewModel()
                    ZettelDetailScreen(
                        zettelId = zettelId,
                        viewModel = viewModel,
                        onBack = { navController.popBackStack() }
                    )
                }
            }
        }
    }
}

sealed class Screen(val route: String, val label: String, val icon: androidx.compose.ui.graphics.vector.ImageVector) {
    object Garden : Screen("garden", "Garden", Icons.Default.Home)
    object Timeline : Screen("timeline", "Timeline", Icons.Default.DateRange)
    object Relations : Screen("relations", "Relations", Icons.Default.Share)
    object Identity : Screen("identity", "Who", Icons.Default.Person)
}
