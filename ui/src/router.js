import { createRouter, createWebHistory } from 'vue-router'
import OverviewPage from './features/dashboard/OverviewPage.vue'
import UsersPage from './features/users/UsersPage.vue'

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/dashboard/overview' },
    { path: '/dashboard/overview', component: OverviewPage },
    { path: '/users', component: UsersPage },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard/overview' },
  ],
})
