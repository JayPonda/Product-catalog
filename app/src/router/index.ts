import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'


const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/categories",
      name: 'categories',
      component: Home,
      meta: {
        layout: 'LoginLayout'
      }
    },
    {
      path: "/login",
      name: 'login',
      component: Home,
      meta: {
        layout: 'LoginLayout'
      }
    },
    {
      path: "/my-products",
      name: 'products',
      component: Home,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/logout",
      name: 'logout',
      component: Home,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/",
      name: 'browse-products',
      component: Home,
      meta: {
        layout: 'LoginLayout'
      }
    },
  ],
})

export default router
