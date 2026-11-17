import { createApp } from 'vue'
import { Loading, Toast } from 'vant'
import 'vant/lib/index.css'
import './styles/tokens.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(router)
;[Loading, Toast].forEach((c) => app.use(c))
app.mount('#app')
