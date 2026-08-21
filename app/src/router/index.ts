import { createRouter, createWebHistory } from 'vue-router'

// poll briefly for an anchor that may not exist until the view has rendered
function waitForElement(selector: string, timeout = 1000) {
  return new Promise<Element | null>((resolve) => {
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
import ModifyProduct from '@/views/ModifyProduct.vue'


const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/categories",
      name: 'categories',
      component: Category,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/categories/add",
      name: 'categories-add',
      component: ModifyCategory,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/products",
      name: 'products',
      component: Product,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/products/add",
      name: 'products-create',
      component: ModifyProduct,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/products/:id/edit",
      name: 'products-modify',
      component: ModifyProduct,
      meta: {
        layout: 'DefaultLayout'
      }
    },
    {
      path: "/",
      name: "home",
      component:  Home,
      meta: {
        layout: 'LoginLayout'
      }
    }
  ],
  scrollBehavior(to, from, savedPosition) {
    if (to.hash) {
      // the destination view may not be mounted yet, so wait for the anchor to appear
      return waitForElement(to.hash).then((el) =>
        el ? { el, behavior: 'smooth' } : false
      )
    }
    return savedPosition ?? { top: 0 }
  },
})

export default router
