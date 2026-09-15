<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import DownloadsView from './views/DownloadsView.vue'
import ListenerView from './views/ListenerView.vue'
import LogsView from './views/LogsView.vue'
import SettingsView from './views/SettingsView.vue'
import ConfirmModal from './components/ConfirmModal.vue'
import AuthWizard from './components/AuthWizard.vue'
import RemoteLogin from './components/RemoteLogin.vue'
import { hasStoredToken, useAuthToken } from './composables/useAuthToken'
import { useConfirmModal } from './composables/useConfirmModal'
import { useUpdater } from './composables/useUpdater'
import {
  ArrowDownToLine,
  ArrowUpRight,
  CheckCircle2,
  LogOut,
  Menu,
  Radio,
  ScrollText,
  Settings2,
  UserCheck,
  X,
  Zap
} from './icons'

const { token, initToken, setToken, clearToken, authHeaders, isWailsRuntime } = useAuthToken()

// Saber ya en el primer pintado si vamos a cargar la aplicación o a pedir el
// token: leer el almacenamiento del navegador es inmediato, así que no hay
// motivo para enseñar la animación de arranque y sustituirla un instante
// después por la pantalla del token.
const startsAuthenticated = hasStoredToken()
const needsRemoteLogin = ref(false)
const remoteLoginError = ref('')

const downloads = ref([])
const listenerItems = ref([])
// Último ID del registro conocido por el servidor. Cambia en cada snapshot y
// es la señal que usa la vista de logs para pedir solo lo nuevo.
const logsSeq = ref(0)
const loading = ref(false)
const message = ref('')
const error = ref('')
const saving = ref(false)
const hydrated = ref(false)
const host = window.location.host
const logoUrl = `${import.meta.env.BASE_URL}telegramdl-android-icon.svg`
const activeView = ref('downloads')
const mobileMenuOpen = ref(false)
const disk = ref(null)

const themeMap = {
  // --- Fila Superior: Sólidos ---
  0: { primary: '#c0514a', secondary: '#c0514a', accent: '#e67d77', bgBase: '#160808', bgTop: '#3a1614', surface: '#220e0d', surfaceLight: '#2d1211', border: '#4a1e1b', borderLight: '#632824', iconBg: '#3a1614', glow: 'rgba(192, 81, 74, 0.15)', textDim: '#a87e7b', gradient: '#c0514a' },
  1: { primary: '#c87a2f', secondary: '#c87a2f', accent: '#e6a567', bgBase: '#161108', bgTop: '#3a2a14', surface: '#22190d', surfaceLight: '#2d2111', border: '#4a351b', borderLight: '#634724', iconBg: '#3a2a14', glow: 'rgba(200, 122, 47, 0.15)', textDim: '#a8927b', gradient: '#c87a2f' },
  2: { primary: '#8b62cf', secondary: '#8b62cf', accent: '#b69df2', bgBase: '#10081a', bgTop: '#26164d', surface: '#170b2e', surfaceLight: '#1f0e3d', border: '#301b5b', borderLight: '#40237a', iconBg: '#26164d', glow: 'rgba(139, 98, 207, 0.15)', textDim: '#8f7ba8', gradient: '#8b62cf' },
  3: { primary: '#479f29', secondary: '#479f29', accent: '#76c859', bgBase: '#081608', bgTop: '#163a14', surface: '#0d220d', surfaceLight: '#112d11', border: '#1b4a1b', borderLight: '#246324', iconBg: '#163a14', glow: 'rgba(71, 159, 41, 0.15)', textDim: '#7ba87b', gradient: '#479f29' },
  4: { primary: '#3fa7b5', secondary: '#3fa7b5', accent: '#7cd1db', bgBase: '#081616', bgTop: '#143a3d', surface: '#0d2222', surfaceLight: '#112d2d', border: '#1b4a4d', borderLight: '#246366', iconBg: '#143a3d', glow: 'rgba(63, 167, 181, 0.15)', textDim: '#7ba8a8', gradient: '#3fa7b5' },
  5: { primary: '#38a7ff', secondary: '#38a7ff', accent: '#5ebcff', bgBase: '#07111f', bgTop: '#163557', surface: '#0b1a2a', surfaceLight: '#0e2032', border: '#1b344b', borderLight: '#234765', iconBg: '#11385b', glow: 'rgba(73, 182, 255, 0.15)', textDim: '#728ba2', gradient: '#38a7ff' },
  6: { primary: '#c04c7d', secondary: '#c04c7d', accent: '#e67dac', bgBase: '#160811', bgTop: '#3a1428', surface: '#220d18', surfaceLight: '#2d111f', border: '#4a1b32', borderLight: '#632442', iconBg: '#3a1428', glow: 'rgba(192, 76, 125, 0.15)', textDim: '#a87b92', gradient: '#c04c7d' },
  7: { primary: '#7d8b99', secondary: '#7d8b99', accent: '#acb8c2', bgBase: '#121416', bgTop: '#282d33', surface: '#1a1e22', surfaceLight: '#22282d', border: '#353d45', borderLight: '#45505a', iconBg: '#282d33', glow: 'rgba(125, 139, 153, 0.15)', textDim: '#888b8e', gradient: '#7d8b99' },

  // --- Fila Inferior: Degradados ---
  8: { primary: '#c0514a', secondary: '#f08c5d', accent: '#e67d77', bgBase: '#160808', bgTop: '#3a1614', surface: '#220e0d', surfaceLight: '#2d1211', border: '#4a1e1b', borderLight: '#632824', iconBg: '#3a1614', glow: 'rgba(192, 81, 74, 0.15)', textDim: '#a87e7b', gradient: 'linear-gradient(135deg, #c0514a 0%, #f08c5d 100%)' },
  9: { primary: '#c87a2f', secondary: '#f2bc42', accent: '#e6a567', bgBase: '#161108', bgTop: '#3a2a14', surface: '#22190d', surfaceLight: '#2d2111', border: '#4a351b', borderLight: '#634724', iconBg: '#3a2a14', glow: 'rgba(200, 122, 47, 0.15)', textDim: '#a8927b', gradient: 'linear-gradient(135deg, #c87a2f 0%, #f2bc42 100%)' },
  10: { primary: '#8b62cf', secondary: '#d66fd6', accent: '#b69df2', bgBase: '#10081a', bgTop: '#26164d', surface: '#170b2e', surfaceLight: '#1f0e3d', border: '#301b5b', borderLight: '#40237a', iconBg: '#26164d', glow: 'rgba(139, 98, 207, 0.15)', textDim: '#8f7ba8', gradient: 'linear-gradient(135deg, #8b62cf 0%, #d66fd6 100%)' },
  11: { primary: '#479f29', secondary: '#a6c450', accent: '#76c859', bgBase: '#081608', bgTop: '#163a14', surface: '#0d220d', surfaceLight: '#112d11', border: '#1b4a1b', borderLight: '#246324', iconBg: '#163a14', glow: 'rgba(71, 159, 41, 0.15)', textDim: '#7ba87b', gradient: 'linear-gradient(135deg, #479f29 0%, #a6c450 100%)' },
  12: { primary: '#3fa7b5', secondary: '#62d4e3', accent: '#7cd1db', bgBase: '#081616', bgTop: '#143a3d', surface: '#0d2222', surfaceLight: '#112d2d', border: '#1b4a4d', borderLight: '#246366', iconBg: '#143a3d', glow: 'rgba(63, 167, 181, 0.15)', textDim: '#7ba8a8', gradient: 'linear-gradient(135deg, #3fa7b5 0%, #62d4e3 100%)' },
  13: { primary: '#38a7ff', secondary: '#b48bf2', accent: '#5ebcff', bgBase: '#07111f', bgTop: '#163557', surface: '#0b1a2a', surfaceLight: '#0e2032', border: '#1b344b', borderLight: '#234765', iconBg: '#11385b', glow: 'rgba(73, 182, 255, 0.15)', textDim: '#728ba2', gradient: 'linear-gradient(135deg, #38a7ff 0%, #b48bf2 100%)' },
  14: { primary: '#c04c7d', secondary: '#f28b7e', accent: '#e67dac', bgBase: '#160811', bgTop: '#3a1428', surface: '#220d18', surfaceLight: '#2d111f', border: '#4a1b32', borderLight: '#632442', iconBg: '#3a1428', glow: 'rgba(192, 76, 125, 0.15)', textDim: '#a87b92', gradient: 'linear-gradient(135deg, #c04c7d 0%, #f28b7e 100%)' },
  15: { primary: '#7d8b99', secondary: '#b0b8c2', accent: '#acb8c2', bgBase: '#121416', bgTop: '#282d33', surface: '#1a1e22', surfaceLight: '#22282d', border: '#353d45', borderLight: '#45505a', iconBg: '#282d33', glow: 'rgba(125, 139, 153, 0.15)', textDim: '#888888', gradient: 'linear-gradient(135deg, #7d8b99 0%, #b0b8c2 100%)' },

  // --- Temas especiales del panel ---
  // Del 16 en adelante ya no son colores de Telegram Premium sino paletas
  // propias, y además de las variables de color encienden una capa decorativa
  // (ver temas-especiales.css) a través de la clase que pone applyTheme en el
  // <html>. El backend guarda el id tal cual, sin validar rango, así que
  // añadir temas aquí no necesita ningún cambio en Go.
  16: { name: 'Twilight Comet', especial: true, primary: '#d1618f', secondary: '#f5b96b', accent: '#b79af5', bgBase: '#0d0a1e', bgTop: '#3a2568', surface: '#160f2d', surfaceLight: '#1d1539', border: '#2f2258', borderLight: '#41307a', iconBg: '#2f2163', glow: 'rgba(192, 81, 140, 0.20)', textDim: '#9086bb', gradient: 'linear-gradient(135deg, #5b45b0 0%, #c0518c 52%, #f0803c 100%)' },
  17: { name: 'Edge Runner UI', especial: true, secondary: '#00e5ff', primary: '#e0ff00', accent: '#00e5ff', bgBase: '#0a0a0c', bgTop: '#14161d', surface: '#16181e', surfaceLight: '#1e2029', border: '#272a33', borderLight: '#3a3e4a', iconBg: '#1f2230', glow: 'rgba(224, 255, 0, 0.18)', textDim: '#a0a5b5', gradient: 'linear-gradient(120deg, #e0ff00 0%, #a8f53c 42%, #00e5ff 100%)' },
  18: { name: 'Rainy Garden', especial: true, secondary: '#8fa4b5', primary: '#e6ecef', accent: '#8fa4b5', bgBase: '#0f1c11', bgTop: '#2d5a27', surface: '#1a3318', surfaceLight: '#224021', border: '#3a5361', borderLight: '#5c768d', iconBg: '#26402c', glow: 'rgba(230, 236, 239, 0.16)', textDim: '#93a79b', gradient: 'linear-gradient(120deg, #e6ecef 0%, #b9c9d4 46%, #8fa4b5 100%)' },
  19: { name: 'Silent Cherry', especial: true, secondary: '#a0e7e5', primary: '#f48fb1', accent: '#35867f', bgBase: '#faf8fc', bgTop: '#ffe9ef', surface: '#ffffff', surfaceLight: '#fff5f8', border: '#f0e3e9', borderLight: '#e3d4dd', iconBg: '#fdebf1', glow: 'rgba(244, 143, 177, 0.25)', textDim: '#8a7f8c', gradient: 'linear-gradient(120deg, #ffb7c5 0%, #f48fb1 52%, #a0e7e5 100%)' },
  20: { name: 'osX', especial: true, secondary: '#64d2ff', primary: '#0a84ff', accent: '#64d2ff', bgBase: '#09090b', bgTop: '#10131a', surface: '#1c1c1e', surfaceLight: '#2c2c2e', border: '#2c2c2e', borderLight: '#3a3a3c', iconBg: '#26262a', glow: 'rgba(10, 132, 255, 0.28)', textDim: '#98989f', gradient: 'linear-gradient(180deg, #3e9bff 0%, #0a84ff 100%)' }
}

const applyLoaderTheme = (colorId) => {
  const theme = themeMap[colorId] || themeMap[5]
  const root = document.documentElement
  root.style.setProperty('--loader-primary', theme.primary)
  root.style.setProperty('--loader-glow', theme.glow)
  localStorage.setItem('tgdl_loader_color', colorId)
}

// Clase que enciende el decorado de cada tema especial. Los colores de Telegram
// (0-15) no tienen entrada aquí y no pintan nada extra: solo cambian variables.
const clasesTemaEspecial = {
  16: 'tema-twilight-comet',
  17: 'tema-edge-runner',
  18: 'tema-rainy-garden',
  19: 'tema-silent-cherry',
  20: 'tema-osx'
}

// Los temas de fondo claro llevan además la clase `tema-claro`, que enciende
// un bloque común de temas-especiales.css. El panel tiene más de cien colores
// de texto fijos escritos para fondo oscuro y hay que pisarlos una sola vez,
// no una por cada tema claro que se añada.
const temasClaros = new Set([19])

const aplicarClaseTema = (colorId) => {
  const root = document.documentElement
  Object.values(clasesTemaEspecial).forEach(clase => root.classList.remove(clase))
  root.classList.remove('tema-claro')
  const clase = clasesTemaEspecial[colorId]
  if (clase) root.classList.add(clase)
  // Number(): el id puede llegar como texto desde el almacenamiento del
  // navegador, y un Set no hace la conversión por su cuenta.
  if (temasClaros.has(Number(colorId))) root.classList.add('tema-claro')
}

const applyTheme = (colorId) => {
  const theme = themeMap[colorId] || themeMap[5]
  const root = document.documentElement
  root.style.setProperty('--user-primary', theme.primary)
  root.style.setProperty('--user-accent', theme.accent)
  root.style.setProperty('--user-bg-base', theme.bgBase)
  root.style.setProperty('--user-bg-top', theme.bgTop)
  root.style.setProperty('--user-surface', theme.surface)
  root.style.setProperty('--user-surface-light', theme.surfaceLight)
  root.style.setProperty('--user-border', theme.border)
  root.style.setProperty('--user-border-light', theme.borderLight)
  root.style.setProperty('--user-icon-bg', theme.iconBg)
  root.style.setProperty('--user-glow', theme.glow)
  root.style.setProperty('--user-text-dim', theme.textDim)
  root.style.setProperty('--user-gradient', theme.gradient)
  aplicarClaseTema(themeMap[colorId] ? colorId : 5)

  // Persistir el color en el objeto de usuario para que index.html lo use en el arranque
  try {
    const user = JSON.parse(localStorage.getItem('tgdl_user') || 'null')
    if (user) {
      user.color_id = colorId
      localStorage.setItem('tgdl_user', JSON.stringify(user))
    }
  } catch (e) {}
}

const storedUser = JSON.parse(localStorage.getItem('tgdl_user') || 'null')
const storedLoaderColor = localStorage.getItem('tgdl_loader_color')

if (storedUser && storedUser.color_id !== undefined) {
  applyTheme(storedUser.color_id)
  if (storedLoaderColor !== null) {
    applyLoaderTheme(parseInt(storedLoaderColor))
  } else {
    applyLoaderTheme(storedUser.loader_color_id !== undefined ? storedUser.loader_color_id : storedUser.color_id)
  }
} else {
  applyTheme(5)
  applyLoaderTheme(storedLoaderColor !== null ? parseInt(storedLoaderColor) : 5)
}

// En un dispositivo nuevo no hay nada guardado, así que la pantalla de acceso
// remoto saldría con el azul por defecto. El color del panel se puede consultar
// sin token (es solo un número, no dice nada de la cuenta), así que lo pedimos
// para que el login ya se vea con el color que el usuario tiene configurado.
const loadPublicTheme = async () => {
  try {
    const response = await fetch('/api/theme')
    if (!response.ok) return
    const data = await response.json()
    if (data.color_id === undefined || data.color_id === null) return
    applyTheme(data.color_id)
    applyLoaderTheme(
      data.loader_color_id === undefined || data.loader_color_id === null
        ? data.color_id
        : data.loader_color_id
    )
  } catch (e) {
    // Sin respuesta del servidor se queda el color por defecto.
  }
}

const authStatus = ref({
  authenticated: localStorage.getItem('tgdl_auth') === 'true',
  state: localStorage.getItem('tgdl_auth') === 'true' ? 'LOGGED_IN' : 'UNCONFIGURED',
  has_credentials: true,
  user: JSON.parse(localStorage.getItem('tgdl_user') || 'null')
})

const settings = reactive({
  max_concurrent_downloads: 2,
  parallel_chunks: true,
  chunk_workers: 8,
  speed_limit: { value: 0, unit: 'MB' },
  color_id: 5,
  loader_color_id: 5,
  download_folder: ''
})

const resetColor = () => {
  if (authStatus.value.user && authStatus.value.user.color_id !== undefined) {
    settings.color_id = authStatus.value.user.color_id
  } else {
    settings.color_id = 5
  }
}

const resetLoaderColor = () => {
  settings.loader_color_id = settings.color_id
}

watch(() => settings.color_id, (newVal) => {
  if (newVal !== undefined) applyTheme(newVal)
})

watch(() => settings.loader_color_id, (newVal) => {
  if (newVal !== undefined) applyLoaderTheme(newVal)
})

let timer
let socket
let reconnectTimer
let saveTimer
let updateCheckTimer
let disposed = false
let syncingSettings = false
const settingsSavePending = ref(false)
const websocketConnected = ref(false)
const bootstrapping = ref(true)
// La animación de arranque solo se muestra cuando de verdad se va a cargar la
// aplicación. Si ya sabemos que toca pedir el token, se espera en silencio a
// conocer el color y se pinta la pantalla del token directamente.
const showSplash = ref(startsAuthenticated)

// Sin token guardado se pide el color antes de pintar la pantalla del token,
// para que no aparezca primero en azul y cambie de color un instante después.
// Si el servidor tardara en responder, se muestra igualmente pasado un margen
// corto en lugar de dejar la pantalla en blanco.
if (!startsAuthenticated) {
  Promise.race([
    loadPublicTheme(),
    new Promise(resolve => setTimeout(resolve, 400))
  ]).finally(() => {
    bootstrapping.value = false
    needsRemoteLogin.value = true
  })
}
const resolvedFileNames = new Map()
const duplicatePrompted = new Set()

const isPlaceholderFileName = name => /^mensaje_\d+$/i.test(String(name || '').trim())
const mergeDownloadNames = items => items.map(item => {
  const currentName = String(item.file_name || '').trim()
  if (currentName && !isPlaceholderFileName(currentName)) {
    resolvedFileNames.set(item.id, currentName)
  } else if (resolvedFileNames.has(item.id)) {
    return { ...item, file_name: resolvedFileNames.get(item.id) }
  }
  return item
})

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const parseSpeed = (speedStr) => {
  if (!speedStr || typeof speedStr !== 'string') return 0
  const match = speedStr.match(/^([\d.]+)\s*([A-Za-z]+)\/s$/)
  if (!match) return 0
  const multipliers = { 'B': 1, 'KB': 1024, 'MB': 1024 ** 2, 'GB': 1024 ** 3, 'TB': 1024 ** 4 }
  return parseFloat(match[1]) * (multipliers[match[2].toUpperCase()] || 1)
}

const syncSettings = async nextSettings => {
  syncingSettings = true
  Object.assign(settings, nextSettings)
  await nextTick()
  syncingSettings = false
}

const showMessage = (txt, isError = false) => {
  if (isError) {
    if (txt.toLowerCase().includes('espacio no es suficiente') || txt.toLowerCase().includes('libera espacio')) {
      openConfirm({
        title: '¡Espacio en disco insuficiente!',
        message: txt,
        confirmText: 'Entendido',
        cancelText: '',
        type: 'danger',
        action: () => {}
      })
      return
    }
    error.value = txt
    setTimeout(() => { if (error.value === txt) error.value = '' }, 5000)
  } else {
    message.value = txt
    setTimeout(() => { if (message.value === txt) message.value = '' }, 3500)
  }
}

const { modal, openConfirm, handleConfirm, handleCancel } = useConfirmModal()

const api = async (url, options = {}) => {
  const response = await fetch(url, {
    ...options,
    headers: { ...(options.headers || {}), ...authHeaders() }
  })
  if (response.status === 401) {
    if (!isWailsRuntime()) {
      clearToken()
      needsRemoteLogin.value = true
      loadPublicTheme()
    }
    throw new Error('No autorizado: token de acceso inválido o ausente')
  }
  const data = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(data.detail || data.error || 'Error en el servidor')
  return data
}

const fetchAuthStatus = async () => {
  try {
    const data = await api('/api/auth/status')
    authStatus.value = data
    if (data.authenticated) {
      localStorage.setItem('tgdl_auth', 'true')
      localStorage.setItem('tgdl_user', JSON.stringify(data.user || null))
      if (data.user && data.user.color_id !== undefined) applyTheme(data.user.color_id)
    } else {
      localStorage.removeItem('tgdl_auth')
      localStorage.removeItem('tgdl_user')
      applyTheme(5)
    }
  } catch (err) {
    console.error('Error verificando estado de sesión:', err)
  }
}

const logoutTelegram = async () => {
  openConfirm({
    title: 'Cerrar sesión de Telegram',
    message: '¿Estás seguro de que deseas cerrar sesión? Tendrás que volver a autenticarte desde la web.',
    confirmText: 'Sí, cerrar sesión',
    type: 'danger',
    action: async () => {
      try {
        await api('/api/auth/logout', { method: 'POST' })
        localStorage.removeItem('tgdl_auth')
        localStorage.removeItem('tgdl_user')
        await fetchAuthStatus()
        showMessage('Sesión cerrada con éxito')
      } catch (err) { showMessage(err.message, true) }
    }
  })
}

const onAuthSuccess = async (newStatus) => {
  authStatus.value = newStatus
  if (newStatus && (newStatus.authenticated || newStatus.state === 'LOGGED_IN')) {
    localStorage.setItem('tgdl_auth', 'true')
    localStorage.setItem('tgdl_user', JSON.stringify(newStatus.user))
    if (newStatus.user && newStatus.user.color_id !== undefined) applyTheme(newStatus.user.color_id)
    await fetchAuthStatus()
    await fetchSettings()
    await fetchDownloads()
  }
}

const fetchDownloads = async () => {
  try {
    const data = await api('/api/downloads')
    if (Array.isArray(data)) {
      const normalizedDownloads = mergeDownloadNames(data)
      downloads.value = normalizedDownloads
      normalizedDownloads.filter(item => item.status === 'duplicate').forEach(promptDuplicateDownload)
      websocketConnected.value = true
    }
    error.value = ''
  } catch (err) { /* No mostramos error en poll constante */ }
}

// El WebSocket (o los eventos nativos de Wails) ya empujan el estado en
// tiempo real; este polling es solo el respaldo para cuando esa conexión
// está caída. Mientras esté sana, lo espaciamos mucho (chequeo de vida)
// en vez de repetir cada segundo la misma información que ya llegó por
// el canal push, que era el mayor costo de red/CPU en historiales grandes.
const scheduleDownloadsPoll = () => {
  if (disposed) return
  fetchDownloads().finally(() => {
    if (disposed) return
    const delay = websocketConnected.value ? 15000 : 1000
    timer = setTimeout(scheduleDownloadsPoll, delay)
  })
}

const fetchSettings = async () => {
  try {
    const data = await api('/api/settings')
    await syncSettings(data.settings)
    hydrated.value = true
  } catch (err) { showMessage(err.message, true) }
}

const saveSettings = async () => {
  saving.value = true
  try {
    const data = await api('/api/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings)
    })
    await syncSettings(data.settings)
    showMessage('Configuración guardada')
  } catch (err) { showMessage(err.message, true) } finally { saving.value = false; settingsSavePending.value = false }
}

const regenerateToken = () => {
  openConfirm({
    title: 'Regenerar token de acceso',
    message: 'El token actual dejará de funcionar de inmediato. Cualquier otro dispositivo (celular, otra PC) que lo esté usando para acceso remoto necesitará que le pases el nuevo token.',
    confirmText: 'Sí, regenerar',
    type: 'danger',
    action: async () => {
      try {
        let newToken = ''
        if (isWailsRuntime() && window.go?.main?.App?.RegenerateLocalToken) {
          newToken = await window.go.main.App.RegenerateLocalToken()
        } else {
          const data = await api('/api/auth/token/regenerate', { method: 'POST' })
          newToken = data.token
        }
        setToken(newToken)
        showMessage('Token regenerado')
      } catch (err) {
        showMessage(err.message, true)
      }
    }
  })
}

const clearDownloadHistory = () => {
  openConfirm({
    title: 'Limpiar estadísticas',
    message: '¿Quieres eliminar del historial todas las descargas completadas, omitidas, fallidas y canceladas? Los archivos del disco no se borrarán.',
    confirmText: 'Sí, limpiar historial',
    type: 'danger',
    action: async () => {
      try {
        const data = await api('/api/downloads/history', { method: 'DELETE' })
        await fetchDownloads()
        showMessage(`${data.removed || 0} registros eliminados`)
      } catch (err) { showMessage(err.message, true) }
    }
  })
}

watch(settings, () => {
  if (!hydrated.value || syncingSettings) return
  settingsSavePending.value = true
  clearTimeout(saveTimer)
  saveTimer = setTimeout(saveSettings, 450)
}, { deep: true })

const startDownload = async (url) => {
  const targetUrl = typeof url === 'string' ? url.trim() : ''
  if (!targetUrl) return
  loading.value = true
  try {
    await api('/api/download', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: targetUrl })
    })
    showMessage('Descarga añadida a la cola')
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) } finally { loading.value = false }
}

const cancelDownload = async (id) => {
  try {
    await api('/api/cancel', {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id })
    })
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) }
}

const setDownloadPause = async (id, paused) => {
  try {
    await api(`/api/${paused ? 'pause' : 'resume'}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id })
    })
    showMessage(paused ? 'Descarga pausada' : 'Descarga reanudada')
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) }
}

const retryDownload = async (item) => {
  try {
    await api('/api/retry', {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id: item.id })
    })
    showMessage('Reintentando descarga…')
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) }
}

const pauseAllDownloads = async () => {
  try {
    await api('/api/downloads/pause-all', { method: 'POST' })
    showMessage('Todas las descargas pausadas')
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) }
}

const resumeAllDownloads = async () => {
  try {
    await api('/api/downloads/resume-all', { method: 'POST' })
    showMessage('Todas las descargas reanudadas')
    await fetchDownloads()
  } catch (err) { showMessage(err.message, true) }
}

const cancelAllDownloads = async () => {
  openConfirm({
    title: 'Cancelar todo',
    message: '¿Estás seguro de que quieres cancelar todas las descargas activas y en cola?',
    confirmText: 'Sí, cancelar todo',
    type: 'danger',
    action: async () => {
      try {
        await api('/api/downloads/cancel-all', { method: 'POST' })
        showMessage('Todas las descargas canceladas')
        await fetchDownloads()
      } catch (err) { showMessage(err.message, true) }
    }
  })
}

const deleteDownload = async item => {
  openConfirm({
    title: 'Borrar archivo',
    message: `¿Estás seguro de que quieres eliminar "${item.file_name}" del servidor? Esta acción no se puede deshacer.`,
    confirmText: 'Sí, borrar archivo',
    type: 'danger',
    action: async () => {
      try {
        await api(`/api/downloads/${encodeURIComponent(item.id)}?delete_file=true`, { method: 'DELETE' })
        showMessage('Archivo borrado')
        await fetchDownloads()
      } catch (err) { showMessage(err.message, true) }
    }
  })
}

const openFile = async (item) => {
  try {
    await api('/api/downloads/open', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: item.id })
    })
  } catch (err) {
    showMessage(err.message, true)
  }
}

const {
  version,
  updateInfo,
  isUpdating,
  isUpdateForced,
  updatePostponedVersion,
  updateProgress,
  installUpdate,
  checkForUpdates,
} = useUpdater({ api, showMessage, openConfirm })

const viewTitle = computed(() => ({
  downloads: 'Descargas',
  listener: 'Escucha',
  logs: 'Logs',
  settings: 'Ajustes'
}[activeView.value] || 'Descargas'))

const speedText = computed(() => settings.speed_limit.value > 0 ? `${settings.speed_limit.value} ${settings.speed_limit.unit}/s` : 'Sin límite')
const totalSpeed = computed(() => {
  const totalBytes = downloads.value.reduce((acc, item) =>
    item.status === 'downloading' ? acc + parseSpeed(item.speed) : acc, 0)
  return totalBytes > 0 ? formatSize(totalBytes) + '/s' : '0 B/s'
})

const wasSpaceCritical = ref(false)
const trackedQueueIds = new Set()
let queueNotificationArmed = false

const isDownloadFinished = status => ['completed', 'skipped', 'failed', 'cancelled'].includes(status)
const isDownloadSuccessful = status => ['completed', 'skipped'].includes(status)

const promptDuplicateDownload = item => {
  if (duplicatePrompted.has(item.id) || modal.show) return

  duplicatePrompted.add(item.id)
  openConfirm({
    title: 'Archivo ya descargado',
    message: `El archivo "${item.file_name}" ya existe en la carpeta de descargas. ¿Quieres descargarlo nuevamente?`,
    confirmText: 'Descargar de nuevo',
    cancelText: 'Cancelar',
    type: 'primary',
    action: async () => {
      try {
        await api('/api/retry', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id: item.id })
        })
        await fetchDownloads()
      } catch (err) {
        duplicatePrompted.delete(item.id)
        showMessage(err.message, true)
      }
    },
    cancelAction: async () => {
      // Cancelar el aviso de duplicado solo descarta la entrada del historial:
      // el archivo existente en disco (que pertenece a la descarga original)
      // no se toca.
      try {
        await api(`/api/downloads/${encodeURIComponent(item.id)}?delete_file=false`, { method: 'DELETE' })
        await fetchDownloads()
      } catch (err) {
        showMessage(err.message, true)
      }
    }
  })
}

const maybeNotifyQueueFinished = currentDownloads => {
  if (!queueNotificationArmed || trackedQueueIds.size === 0) return

  const byID = new Map(currentDownloads.map(item => [item.id, item]))
  for (const id of trackedQueueIds) {
    if (!byID.has(id)) trackedQueueIds.delete(id)
  }
  const trackedItems = [...trackedQueueIds].map(id => byID.get(id)).filter(Boolean)
  if (trackedItems.length === 0) {
    queueNotificationArmed = false
    return
  }
  if (!trackedItems.every(item => isDownloadFinished(item.status))) return

  const successful = trackedItems.every(item => isDownloadSuccessful(item.status))
  trackedQueueIds.clear()
  queueNotificationArmed = false

  if (successful) {
    new Audio(`${import.meta.env.BASE_URL}notification.wav`).play().catch(e => console.log('Audio blocked', e))
  }
}

watch(() => disk.value, (newDisk) => {
  if (!newDisk) return
  if (newDisk.projected_free < 0) {
    if (!wasSpaceCritical.value && !modal.show) {
      wasSpaceCritical.value = true
      const needed = formatSize(Math.abs(newDisk.projected_free))
      openConfirm({
        title: '¡Alerta de Espacio Crítico!',
        message: `Debido a cambios externos en tu disco, ya no hay espacio suficiente para completar las descargas en cola. \n\nNecesitas liberar al menos ${needed} o cancelar algunas tareas para evitar errores.`,
        confirmText: 'Entendido',
        cancelText: '',
        type: 'danger',
        action: () => {}
      })
    }
  } else {
    wasSpaceCritical.value = false
  }
}, { deep: true })

const handleStateUpdate = (data) => {
  if (!data) return
  if (Array.isArray(data.downloads)) {
    const normalizedDownloads = mergeDownloadNames(data.downloads)
    downloads.value = normalizedDownloads
    normalizedDownloads.filter(item => item.status === 'duplicate').forEach(promptDuplicateDownload)
    data.downloads.forEach(item => {
      if (['queued', 'downloading'].includes(item.status)) {
        trackedQueueIds.add(item.id)
        queueNotificationArmed = true
      }
    })
    maybeNotifyQueueFinished(data.downloads)
  }
  if (Array.isArray(data.listener)) {
    listenerItems.value = data.listener
  }
  if (typeof data.logs_seq === 'number') {
    logsSeq.value = data.logs_seq
  }
  if (data.disk) {
    disk.value = data.disk
  }
  if (data.settings && !settingsSavePending.value && !saving.value) syncSettings(data.settings)
}

const connectWebSocket = async () => {
  if (window.runtime && window.runtime.EventsOn) {
    websocketConnected.value = true
    window.runtime.EventsOn("tgdl:state", handleStateUpdate)
  }

  let wsUrl = ''
  if (window.location.host && !window.location.host.includes('wails')) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    wsUrl = `${proto}//${window.location.host}/api/ws`
  } else {
    wsUrl = 'ws://127.0.0.1:8000/api/ws'
  }
  // El token viaja como subprotocolo, no en la URL. Las URL completas quedan
  // registradas en cualquier proxy o túnel por el que pase la conexión, que es
  // justo lo que se usa para el acceso remoto; la cabecera del subprotocolo no.
  const protocolos = ['tgdl-v1']
  if (token.value) {
    protocolos.push(token.value)
  }

  try {
    socket = new WebSocket(wsUrl, protocolos)
    socket.onopen = () => { websocketConnected.value = true }
    socket.onmessage = event => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'state') {
          handleStateUpdate(data)
        }
      } catch {}
    }
    socket.onclose = () => {
      if (!window.runtime) websocketConnected.value = false
      if (!disposed) reconnectTimer = setTimeout(connectWebSocket, 2000)
    }
    socket.onerror = () => {
      socket?.close()
    }
  } catch {
    if (!window.runtime) websocketConnected.value = false
    if (!disposed) reconnectTimer = setTimeout(connectWebSocket, 2000)
  }
}

const startApp = async () => {
  disposed = false

  // Iniciar servicios y comprobación inmediata
  connectWebSocket()
  scheduleDownloadsPoll()

  // Iniciamos el ciclo de actualizaciones de fondo inmediatamente para evitar esperas
  setTimeout(() => checkForUpdates(false), 1000)
  // Cada 5 minutos: la API de GitHub solo admite 60 peticiones por hora y por
  // IP, y a 2 minutos se agotaba la cuota sin necesidad.
  updateCheckTimer = setInterval(() => checkForUpdates(false), 5 * 60 * 1000)

  const startBoot = Date.now()
  try {
    await Promise.all([fetchAuthStatus(), checkForUpdates(true)])
    if (authStatus.value.authenticated) {
      await Promise.all([fetchSettings(), fetchDownloads()]).catch(() => {})
    }
  } catch (err) {
    console.error('Error durante el arranque:', err)
  } finally {
    const elapsed = Date.now() - startBoot
    const minTime = 6500 // 4.5s animación + 2s de reposo con puntos
    const remaining = Math.max(0, minTime - elapsed)
    setTimeout(() => {
      bootstrapping.value = false
    }, remaining)
  }
}

// Valida un token recién introducido en la pantalla de login remoto antes de
// arrancar el resto de la app con él.
const handleRemoteLogin = async (value) => {
  remoteLoginError.value = ''
  setToken(value)
  try {
    await api('/api/system/info')
    needsRemoteLogin.value = false
    // Token aceptado: a partir de aquí sí se carga la aplicación, así que la
    // animación de arranque vuelve a tener sentido.
    showSplash.value = true
    bootstrapping.value = true
    await startApp()
  } catch (err) {
    clearToken()
    remoteLoginError.value = 'Token inválido. Verifica que lo copiaste completo desde Ajustes → Acceso remoto.'
  }
}

onMounted(async () => {
  // Resolver el token de acceso antes de llamar a cualquier endpoint /api/*:
  // en la app de escritorio llega solo (binding nativo de Wails); en un
  // navegador remoto, si no hay uno guardado, pedimos la pantalla de login.
  await initToken()
  if (isWailsRuntime() || token.value) {
    await startApp()
  } else {
    bootstrapping.value = false
    needsRemoteLogin.value = true
    // Solo si creíamos tener token y resultó no haberlo: en el otro caso el
    // color ya se pidió antes de pintar nada.
    if (startsAuthenticated) {
      await loadPublicTheme()
    }
  }
})

onUnmounted(() => {
  disposed = true
  clearTimeout(timer)
  clearInterval(updateCheckTimer)
  clearTimeout(saveTimer)
  clearTimeout(reconnectTimer)
  socket?.close()
})
</script>

<template>
  <!-- Pantalla de arranque -->
  <transition name="fade">
    <div v-if="bootstrapping && showSplash" class="boot-screen">
      <div class="boot-container">
        <svg viewBox="0 0 500 150" class="hello-svg">
          <text x="50%" y="50%" text-anchor="middle" dominant-baseline="middle" class="hello-text">
            <tspan class="c1">T</tspan><tspan class="c2">e</tspan><tspan class="c3">l</tspan><tspan class="c4">e</tspan><tspan class="c5">g</tspan><tspan class="c6">r</tspan><tspan class="c7">a</tspan><tspan class="c8">m</tspan><tspan class="c9">D</tspan><tspan class="c10">L</tspan>
            <tspan class="dots" dx="20" dy="8">
              <tspan class="dot1">●</tspan><tspan class="dot2">●</tspan><tspan class="dot3">●</tspan>
            </tspan>
          </text>
        </svg>
      </div>
    </div>
  </transition>

  <template v-if="!bootstrapping">
    <!-- Diálogo de actualización obligatoria -->
    <div v-if="updateInfo && isUpdateForced" class="update-required-overlay">
      <div class="update-card">
        <Zap :size="48" class="update-icon" />
        <h2>Actualización Obligatoria</h2>
        <p v-if="!isUpdating">Hay una nueva versión disponible ({{ updateInfo.latest }}). Es necesario actualizar para continuar.</p>

        <div class="update-action-area" :class="{ 'is-loading': isUpdating }">
          <button v-if="!isUpdating" class="primary-button update-btn" @click="installUpdate">
            <span>Actualizar ahora</span>
            <ArrowUpRight :size="18" />
          </button>

          <div v-else class="update-progress-container">
            <div class="update-status-text">
              {{ updateProgress.status === 'downloading' ? 'Descargando actualización...' :
                 updateProgress.status === 'extracting' ? 'Extrayendo archivos...' :
                 updateProgress.status === 'finishing' ? 'Finalizando e iniciando...' : 'Iniciando...' }}
            </div>
            <div class="update-progress-bar">
              <div class="update-progress-fill" :style="{ width: updateProgress.percentage + '%' }">
                <div class="nitro-wind">
                  <span></span><span></span><span></span><span></span>
                </div>
              </div>
            </div>
            <div class="update-progress-stats">
              <span>{{ formatSize(updateProgress.downloaded) }} / {{ formatSize(updateProgress.total) }}</span>
              <span>{{ updateProgress.percentage }}%</span>
            </div>
          </div>
        </div>

        <div v-if="isUpdating" class="update-warning">
          Por favor, no cierres la aplicación.
        </div>

        <small>Versión actual: {{ updateInfo.current }}</small>
      </div>
    </div>

    <!-- Login remoto: token de acceso a la API (solo en navegador/remoto sin token guardado) -->
    <RemoteLogin v-if="needsRemoteLogin" :error="remoteLoginError" @submit="handleRemoteLogin" />

    <!-- Asistente de Autenticación -->
    <AuthWizard v-else-if="!authStatus.authenticated" :authStatus="authStatus" @auth-success="onAuthSuccess" />

    <!-- Estructura Principal de la Aplicación -->
    <div v-else class="app-shell">
      <aside class="sidebar" :class="{ 'mobile-open': mobileMenuOpen }">
        <div class="sidebar-main">
          <div class="brand">
            <span class="brand-mark"><img :src="logoUrl" alt="" /></span>
            <div class="brand-text">
              <svg viewBox="0 0 130 45" class="brand-svg">
                <!-- Nombre de la App animado letra a letra -->
                <text x="0" y="18" class="brand-logo-text">
                  <tspan class="b1">T</tspan><tspan class="b2">e</tspan><tspan class="b3">l</tspan><tspan class="b4">e</tspan><tspan class="b5">g</tspan><tspan class="b6">r</tspan><tspan class="b7">a</tspan><tspan class="b8">m</tspan><tspan class="brand-accent b9">D</tspan><tspan class="brand-accent b10">L</tspan>
                </text>
                <!-- Versión animada sincronizada con el texto de arriba -->
                <text x="0" y="38" class="brand-version-text" v-if="version">
                  <tspan v-for="(char, i) in ('v' + version).split('')" :key="i" :class="'b' + (i+1)">{{ char }}</tspan>
                </text>
              </svg>
            </div>
            <button
              class="mobile-menu-toggle"
              type="button"
              :aria-expanded="mobileMenuOpen"
              aria-label="Abrir menú"
              @click="mobileMenuOpen = !mobileMenuOpen"
            >
              <X v-if="mobileMenuOpen" :size="20" />
              <Menu v-else :size="20" />
            </button>
          </div>
          <p class="sidebar-copy">Centro de descargas personal</p>
          <nav class="sidebar-nav">
            <button
              :class="{ selected: activeView === 'downloads' }"
              @click="activeView = 'downloads'; mobileMenuOpen = false"
            >
              <ArrowDownToLine :size="16" /> Descargas
            </button>
            <button
              :class="{ selected: activeView === 'listener' }"
              @click="activeView = 'listener'; mobileMenuOpen = false"
            >
              <Radio :size="16" /> Escucha
            </button>
            <button
              :class="{ selected: activeView === 'logs' }"
              @click="activeView = 'logs'; mobileMenuOpen = false"
            >
              <ScrollText :size="16" /> Logs
            </button>
            <button
              :class="{ selected: activeView === 'settings' }"
              @click="activeView = 'settings'; mobileMenuOpen = false"
            >
              <Settings2 :size="16" /> Ajustes
            </button>
          </nav>

          <div v-if="authStatus.user" class="sidebar-user-badge">
            <div class="user-info">
              <UserCheck :size="14" class="user-icon" />
              <span class="user-name">{{ authStatus.user.first_name }}</span>
            </div>
            <button class="logout-btn" title="Cerrar sesión de Telegram" @click="logoutTelegram">
              <LogOut :size="13" />
            </button>
          </div>

          <div class="sidebar-status" :class="{ 'is-disconnected': !websocketConnected }">
            <span class="status-dot" :class="{ 'disconnected': !websocketConnected }"></span>
            <span>{{ websocketConnected ? 'Servicio conectado' : 'Servicio desconectado' }}</span>
          </div>
        </div>
        <div class="sidebar-bottom">
          <span class="mini-label">LÍMITE ACTUAL</span>
          <strong>{{ settings.max_concurrent_downloads }} descargas</strong>
          <span>{{ speedText }}</span>
        </div>
      </aside>

      <main class="main-content">
        <header class="topbar">
          <div>
            <span class="eyebrow">PANEL DE CONTROL</span>
            <h1>{{ viewTitle }}</h1>
          </div>
          <div class="topbar-meta">Velocidad total: {{ totalSpeed }}</div>
        </header>

        <div v-if="message" class="toast success"><CheckCircle2 :size="15" /> {{ message }}</div>
        <div v-if="error" class="toast danger">{{ error }}</div>

        <!-- Contenedor Modular de Vistas -->
        <div class="view-container">
          <DownloadsView
            v-show="activeView === 'downloads'"
            :downloads="downloads"
            :disk="disk"
            :settings="settings"
            :loading="loading"
            @start-download="startDownload"
            @pause-download="id => setDownloadPause(id, true)"
            @resume-download="id => setDownloadPause(id, false)"
            @cancel-download="cancelDownload"
            @retry-download="retryDownload"
            @delete-download="deleteDownload"
            @open-file="openFile"
            @pause-all="pauseAllDownloads"
            @resume-all="resumeAllDownloads"
            @cancel-all="cancelAllDownloads"
          />

          <ListenerView
            v-show="activeView === 'listener'"
            :notify="showMessage"
            :disk="disk"
            :initialItems="listenerItems"
            :settings="settings"
          />

          <LogsView
            v-show="activeView === 'logs'"
            :notify="showMessage"
            :logsSeq="logsSeq"
            :active="activeView === 'logs'"
          />

          <SettingsView
            v-show="activeView === 'settings'"
            :settings="settings"
            :saving="saving"
            :themeMap="themeMap"
            :api-token="token"
            :notify="showMessage"
            @save-settings="saveSettings"
            @clear-history="clearDownloadHistory"
            @reset-color="resetColor"
            @reset-loader-color="resetLoaderColor"
            @regenerate-token="regenerateToken"
          />
        </div>

        <ConfirmModal
          :show="modal.show"
          :title="modal.title"
          :message="modal.message"
          :confirmText="modal.confirmText"
          :cancelText="modal.cancelText"
          :type="modal.type"
          @confirm="handleConfirm"
          @cancel="handleCancel"
        />

        <footer>TelegramDL · Configuración persistida localmente en SQLite · {{ host }}</footer>
      </main>
    </div>
  </template>
</template>

<style>
@font-face {
  font-family: 'LoaderFont';
  src: url('/fonts/loader.ttf') format('truetype'),
       url('/fonts/loader.otf') format('opentype');
  font-weight: normal;
  font-style: normal;
}

.boot-screen {
  position: fixed;
  inset: 0;
  background: var(--user-bg-base);
  background-image: radial-gradient(circle at 70% -10%, var(--user-bg-top) 0, var(--user-bg-base) 40%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}
.boot-container {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
.hello-svg {
  width: 90%;
  max-width: 600px;
  overflow: visible;
}
.hello-text {
  font-family: 'LoaderFont', cursive;
  font-weight: 700;
  font-size: 86px;
  filter: drop-shadow(0 0 25px var(--loader-glow));
}

.hello-text tspan:not(.dots):not(.dot1):not(.dot2):not(.dot3) {
  fill: var(--loader-primary);
  fill-opacity: 0;
  stroke: var(--loader-primary);
  stroke-width: 1.2;
  stroke-dasharray: 400;
  stroke-dashoffset: 400;
  animation: writeAndFill 1.8s cubic-bezier(0.445, 0.05, 0.55, 0.95) forwards;
}

/* Retrasos escalonados equilibrados para Dancing Script */
.c1 { animation-delay: 0.1s; }
.c2 { animation-delay: 0.5s; }
.c3 { animation-delay: 0.9s; }
.c4 { animation-delay: 1.3s; }
.c5 { animation-delay: 1.7s; }
.c6 { animation-delay: 2.1s; }
.c7 { animation-delay: 2.5s; }
.c8 { animation-delay: 2.9s; }
.c9 { animation-delay: 3.3s; }
.c10 { animation-delay: 3.7s; }

.dots {
  font-family: Arial, sans-serif;
  font-weight: bold;
  font-size: 24px;
  stroke: none !important;
  stroke-width: 0 !important;
  vertical-align: middle;
}

.dot1, .dot2, .dot3 {
  opacity: 0;
  fill: var(--loader-primary) !important;
  animation: dotFade 1.5s infinite;
}

.dot1 { animation-delay: 4.2s; }
.dot2 { animation-delay: 4.4s; }
.dot3 { animation-delay: 4.6s; }

@keyframes dotFade {
  0%, 100% { opacity: 0; }
  50% { opacity: 1; }
}

@keyframes writeAndFill {
  0% {
    stroke-dashoffset: 400;
    fill-opacity: 0;
  }
  40% {
    fill-opacity: 0.3;
  }
  100% {
    stroke-dashoffset: 0;
    fill-opacity: 1;
  }
}

@keyframes write {
  to { stroke-dashoffset: 0; }
}

@keyframes fillText {
  from { fill: transparent; }
  to { fill: var(--user-primary); }
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.8s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

/* Animación del Logo en el Sidebar */
.brand-svg {
  width: 120px;
  height: 40px;
  overflow: visible;
  display: block;
}
.brand-logo-text {
  font-family: 'Space Grotesk', sans-serif;
  font-weight: 700;
  font-size: 19px;
  fill: #f3f8ff;
}
.brand-version-text {
  font-family: 'DM Sans', sans-serif;
  font-size: 11px;
  fill: var(--user-text-dim);
  font-weight: 500;
  letter-spacing: 0.05em;
}
.brand-logo-text .brand-accent {
  fill: var(--user-accent);
}
.brand-logo-text tspan, .brand-version-text tspan {
  opacity: 0;
  fill-opacity: 0;
  stroke: #f3f8ff;
  stroke-width: 0.4;
  stroke-dasharray: 80;
  stroke-dashoffset: 80;
  display: inline-block;
  animation: brandSweepWritingLoop 8s ease-in-out infinite;
}
.brand-version-text tspan {
  stroke: var(--user-text-dim);
  fill: var(--user-text-dim);
}
.brand-logo-text .brand-accent {
  stroke: var(--user-accent);
}

/* Stagger compartido: la letra N del nombre y la letra N de la versión aparecen juntas */
.b1 { animation-delay: 1.0s; }
.b2 { animation-delay: 1.12s; }
.b3 { animation-delay: 1.24s; }
.b4 { animation-delay: 1.36s; }
.b5 { animation-delay: 1.48s; }
.b6 { animation-delay: 1.6s; }
.b7 { animation-delay: 1.72s; }
.b8 { animation-delay: 1.84s; }
.b9 { animation-delay: 1.96s; }
.b10 { animation-delay: 2.08s; }

@keyframes brandSweepWritingLoop {
  0% {
    opacity: 0;
    fill-opacity: 0;
    stroke-dashoffset: 80;
    transform: translateX(-15px);
  }
  25% { /* Entrada: se escribe y se llena mientras se desliza */
    opacity: 1;
    fill-opacity: 1;
    stroke-dashoffset: 0;
    transform: translateX(0);
  }
  75% { /* Pausa: texto fijo y sólido */
    opacity: 1;
    fill-opacity: 1;
    stroke-dashoffset: 0;
    transform: translateX(0);
  }
  90% { /* Salida: se desliza a la derecha y se desvanece */
    opacity: 0;
    fill-opacity: 0;
    transform: translateX(15px);
  }
  100% {
    opacity: 0;
  }
}

.update-required-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: color-mix(in srgb, var(--user-bg-base), transparent 5%);
  backdrop-filter: blur(8px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.update-card {
  background: var(--user-surface);
  border: 1px solid var(--user-border);
  border-radius: 24px;
  padding: 40px;
  max-width: 400px;
  text-align: center;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}
.update-icon { color: var(--user-primary); margin-bottom: 20px; }
.update-card h2 { color: #f8fafc; margin-bottom: 12px; font-size: 24px; }
.update-card p { color: var(--user-text-dim); margin-bottom: 30px; line-height: 1.6; }
.update-action-area { transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1); min-height: 80px; display: flex; flex-direction: column; justify-content: center; position: relative; }
.update-action-area.is-loading { transform: translateY(-5px); }
.update-progress-container { width: 100%; text-align: left; animation: morphReveal 0.8s cubic-bezier(0.19, 1, 0.22, 1); }
.update-status-text { color: #f8fafc; font-size: 14px; margin-bottom: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
.update-progress-bar { height: 16px; background: var(--user-bg-base); border-radius: 20px; overflow: hidden; margin-bottom: 10px; position: relative; border: 1px solid var(--user-border); box-shadow: inset 0 2px 8px rgba(0,0,0,0.5); }
.update-progress-fill { height: 100%; background: linear-gradient(90deg, var(--user-bg-top), var(--user-primary), var(--user-accent)); transition: width 0.4s cubic-bezier(0.1, 0.7, 0.1, 1); position: relative; display: flex; align-items: center; justify-content: flex-end; }

/* Efecto Nitro-Wind */
.nitro-wind { position: absolute; top: 0; left: 0; right: 0; bottom: 0; overflow: hidden; pointer-events: none; }
.nitro-wind span { position: absolute; background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.6), transparent); height: 2px; border-radius: 2px; animation: windBackwards 0.6s linear infinite; }
.nitro-wind span:nth-child(1) { top: 25%; width: 40px; animation-duration: 0.4s; }
.nitro-wind span:nth-child(2) { top: 50%; width: 60px; animation-duration: 0.7s; animation-delay: 0.1s; }
.nitro-wind span:nth-child(3) { top: 75%; width: 30px; animation-duration: 0.5s; animation-delay: 0.2s; }
.nitro-wind span:nth-child(4) { top: 40%; width: 50px; animation-duration: 0.8s; animation-delay: 0.3s; }

.update-progress-stats { display: flex; justify-content: space-between; color: var(--user-text-dim); font-size: 12px; font-family: 'Space Grotesk', monospace; font-weight: 600; }
.update-warning { color: #f87171; font-size: 13px; margin-top: 15px; font-weight: 500; text-align: center; animation: pulseWarning 2s infinite; }

.update-btn {
  width: fit-content !important;
  min-width: 240px !important;
  height: 60px !important;
  padding: 0 40px !important;
  font-size: 17px !important;
  font-weight: 700 !important;
  margin: 0 auto !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  border-radius: 30px !important;
  gap: 12px !important;
  background: linear-gradient(135deg, var(--user-primary) 0%, var(--user-bg-top) 100%) !important;
  color: white !important;
  border: none !important;
  cursor: pointer !important;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275) !important;
}
.update-btn:hover { transform: scale(1.05) translateY(-3px) !important; box-shadow: 0 15px 30px var(--user-glow) !important; }
.update-btn:active { transform: scale(0.98) !important; }

@keyframes windBackwards {
  from { transform: translateX(300px); opacity: 0; }
  50% { opacity: 1; }
  to { transform: translateX(-100px); opacity: 0; }
}
@keyframes morphReveal {
  from { opacity: 0; transform: scaleX(0.5); filter: blur(5px); }
  to { opacity: 1; transform: scaleX(1); filter: blur(0); }
}
.update-card small { color: var(--user-text-dim); opacity: 0.8; display: block; }

@keyframes nitro {
  from { transform: translateX(-100%); }
  to { transform: translateX(100%); }
}
@keyframes morphIn {
  from { opacity: 0; transform: translateY(10px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes pulseWarning {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
.sidebar-nav{display:flex;flex-direction:column;gap:6px;margin-bottom:24px}
.sidebar-user-badge{display:flex;align-items:center;justify-content:space-between;background:var(--user-bg-base);border:1px solid var(--user-border);border-radius:10px;padding:8px 10px;margin-bottom:14px;font-size:12px;color:#dbe7f5}
.sidebar-user-badge .user-info{display:flex;align-items:center;gap:6px;min-width:0;overflow:hidden}
.sidebar-user-badge .user-icon{color:#39db9a;flex:none}
.sidebar-user-badge .user-name{white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-weight:600}
.sidebar-user-badge .logout-btn{background:transparent;border:0;color:#e88888;cursor:pointer;display:grid;place-items:center;padding:4px;border-radius:6px;flex:none}
.sidebar-user-badge .logout-btn:hover{background:#3d171d;color:#ff9e9e}

/* Estilos para Ajustes y Selector de Color */
.content-grid-single { display: grid; grid-template-columns: 1fr; gap: 18px; animation: riseIn .55s ease both; }
.settings-panel-full { padding: 30px; }
.settings-sections-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 40px; margin-bottom: 20px; align-items: start; }
.settings-group { display: flex; flex-direction: column; gap: 15px; }
.settings-actions { display: flex; gap: 12px; margin-top: 22px; }
.settings-actions .clear-history-button, .settings-actions .save-button { width: 100%; margin-top: 0; flex: 1; }
.clear-history-button { width: 100%; margin-top: 22px; padding: 11px; border: 1px solid #6e3942; border-radius: 10px; background: rgba(125, 48, 61, .16); color: #ffadb5; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 12px; }
.clear-history-button:hover { background: rgba(125, 48, 61, .3); border-color: #a95663; }
/* La columna de colores es la mas alta del bloque; ocupando dos filas deja
   que el ultimo grupo (Acceso remoto) suba al hueco de la fila 2 en vez de
   quedar colgado debajo. Solo se nota donde no caben las 4 columnas. */
.color-group { padding-top: 5px; grid-row: span 2; }
.color-selector-container { display: flex; flex-direction: column; gap: 12px; margin: 10px 0 20px; }
.color-row { display: flex; gap: 12px; flex-wrap: wrap; }
.color-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  padding: 0;
  transition: all 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  position: relative;
  overflow: hidden;
}
.color-dot:hover { transform: scale(1.15); }
.color-dot.active {
  transform: scale(1.1);
  box-shadow:
    0 0 0 3px var(--user-bg-base),
    0 0 0 5px var(--user-primary),
    0 0 15px var(--user-glow);
  z-index: 2;
}
.gradient-dot { position: relative; }

/* Colores temáticos (ids 16+). Se presentan con el nombre a la vista y no como
   un punto más: son pocos, no vienen de Telegram y el nombre es lo que los
   distingue. */
.seccion-temas { margin-top: 30px; border-top: 1px solid var(--user-border); padding-top: 20px; }
.temas-nota { font-size: 11px; color: var(--user-text-dim); line-height: 1.5; margin-top: -6px; }
.temas-lista { display: flex; flex-wrap: wrap; gap: 10px; margin: 4px 0 20px; }
.tema-chip {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 7px 13px 7px 8px;
  border-radius: 999px;
  border: 1px solid var(--user-border);
  background: var(--user-surface-light);
  color: var(--user-text-dim);
  font: 600 12px 'DM Sans';
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.tema-chip:hover { transform: translateY(-2px); border-color: var(--user-border-light); color: #eef7ff; }
.tema-chip.active {
  color: #eef7ff;
  border-color: var(--user-primary);
  box-shadow: 0 0 0 1px var(--user-primary), 0 6px 18px var(--user-glow);
}
.tema-muestra {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  flex: none;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
}
.reset-button-alt {
  background: var(--user-surface-light);
  border: 1px solid var(--user-border);
  color: var(--user-text-dim);
  padding: 10px 16px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
  transition: all 0.2s;
}
.reset-button-alt:hover {
  background: var(--user-icon-bg);
  color: var(--user-accent);
  border-color: var(--user-primary);
}

.action-btn-mini {
  background: var(--user-surface-light);
  border: 1px solid var(--user-border);
  color: var(--user-text-dim);
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}
.action-btn-mini:hover {
  background: var(--user-icon-bg);
  color: var(--user-accent);
  border-color: var(--user-primary);
}
.action-btn-mini.danger:hover {
  background: rgba(125, 48, 61, .3);
  border-color: #a95663;
  color: #ffadb5;
}
</style>
