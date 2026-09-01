import './assets/main.css'
import 'notivue/notifications.css'
import 'notivue/animations.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createNotivue } from 'notivue'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(
  createNotivue({
    position: 'top-center',
    limit: 4,
  }),
)

app.mount('#app')
