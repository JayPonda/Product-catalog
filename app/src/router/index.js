import { createRouter, createWebHistory } from 'vue-router'
import logger from '@/utils/logger'

// poll briefly for an anchor that may not exist until the view has rendered
function waitForElement(selector, timeout = 1000) {
  return new Promise((resolve) => {
    if (typeof document === 'undefined') return resolve(null)
    const found = document.querySelector(selector)
    if (found) return resolve(found)
    const start = Date.now()
    const timer = setInterval(() => {
      const el = document.querySelector(selector)
      if (el) {
        clearInterval(timer)
        resolve(el)
      } else if (Date.now() - start > timeout) {
        clearInterval(timer)
        resolve(null)
      }
    }, 50)
  })
}

import Home from '@/views/Home.vue'
import Category from '@/views/Category.vue'
import ModifyCategory from '@/components/Category/ModifyCategory.vue'
import Product from '@/views/Product.vue'
import MyProducts from '@/views/MyProducts.vue'
import ModifyProduct from '@/views/ModifyProduct.vue'
import Login from '@/views/Login.vue'
import Register from '@/views/Register.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/categories',
      name: 'categories',
      component: Category,
    },
    {
      path: '/categories/add',
      name: 'categories-add',
      component: ModifyCategory,
    },
    {
      path: '/products',
      name: 'products',
      component: Product,
    },
    {
      path: '/my-products',
      name: 'my-products',
      component: MyProducts,
    },
    {
      path: '/products/add',
      name: 'products-create',
      component: ModifyProduct,
    },
    {
      path: '/products/:id/edit',
      name: 'products-modify',
      component: ModifyProduct,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/register',
      name: 'register',
      component: Register,
    },
    {
      path: '/',
      name: 'home',
      component: Home,
    },
  ],
  scrollBehavior(to, from, savedPosition) {
    if (to.hash) {
      // the destination view may not be mounted yet, so wait for the anchor to appear
      return waitForElement(to.hash).then((el) => (el ? { el, behavior: 'smooth' } : false))
    }
    return savedPosition ?? { top: 0 }
  },
})

// Redirect unauthenticated users away from protected routes.
let authInitialized = false
router.beforeEach(async (to) => {
  logger.Debug('router/index.js', 'beforeEach', `navigating to ${to.path}`, { name: to.name })

  const auth = useAuthStore()
  if (!authInitialized) {
    await auth.fetchMe()
    authInitialized = true
  }

  const isProtected =
    to.path === '/products/add' ||
    to.path === '/my-products' ||
    to.path === '/categories/add' ||
    /^\/products\/[^/]+\/edit$/.test(to.path)

  if (isProtected && !auth.isAuthenticated) {
    logger.Warn('router/index.js', 'beforeEach', 'unauthenticated access to protected route', { path: to.path })
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
