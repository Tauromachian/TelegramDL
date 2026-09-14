import { createApp } from 'vue'
import App from './App.vue'
import './style.css'
// Decorado de los temas especiales (color 16 en adelante). Va todo dentro de
// una clase que solo se pone en el <html> cuando uno de esos temas está
// activo, así que con los colores de Telegram no hay ni una regla aplicada.
import './temas-especiales.css'
import { isWailsRuntime } from './composables/useAuthToken'

// Dentro de la app de escritorio el panel debe comportarse como una ventana
// nativa y no como una página web (ver .app-nativo en style.css). Se comprueba
// dos veces por si el bundle llegara a ejecutarse antes de que Wails inyecte
// sus bindings: la segunda pasada, ya con la página cargada, no falla nunca.
const marcarNativo = () => {
  if (isWailsRuntime()) document.documentElement.classList.add('app-nativo')
}
marcarNativo()
window.addEventListener('load', marcarNativo, { once: true })

createApp(App).mount('#app')
