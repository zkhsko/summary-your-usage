import { createRouter, createWebHistory } from 'vue-router'
import OverviewPage from './features/dashboard/OverviewPage.vue'
import UsersPage from './features/users/UsersPage.vue'
import GroupsPage from './features/groups/GroupsPage.vue'

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/dashboard/overview' },
    { path: '/dashboard/overview', component: OverviewPage },
    { path: '/users', component: UsersPage },
    { path: '/groups', component: GroupsPage },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard/overview' },
  ],
})
