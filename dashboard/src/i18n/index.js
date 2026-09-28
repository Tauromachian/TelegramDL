// Infraestructura de idiomas del panel.
//
// Estado global (fuera de cualquier componente) para que todas las vistas y
// los composables lean el mismo idioma y reaccionen al cambio. es.js es la
// fuente de verdad y también el fallback: cualquier clave que falte en el
// idioma activo se resuelve en español antes de darse por perdida.
//
// Persistencia en dos capas, igual que el color del tema:
//  - localStorage ('tgdl_lang'): caché del dispositivo. Es lo que se lee antes
//    del primer pintado, cuando todavía no se ha podido pedir nada al servidor
//    (la pantalla de acceso remoto y el asistente de Telegram se ven antes de
//    tener token).
//  - settings.language en el servidor (SQLite): la elección definitiva, que
//    App.vue aplica en cuanto llegan los ajustes.
//
// La primera vez que se abre el panel en un dispositivo no hay nada guardado,
// así que se detecta el idioma del sistema: español si el navegador/equipo
// está en español, inglés en cualquier otro caso.
import { computed, reactive } from 'vue'
import es from './es.js'
import en from './en.js'

const dictionaries = { es, en }

// Idiomas que se pueden elegir en Ajustes. La etiqueta va en su propio idioma
// a propósito: quien busca "English" tiene que reconocerlo aunque el panel
// esté ahora mismo en español.
export const availableLocales = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English' }
]

const STORAGE_KEY = 'tgdl_lang'

const isValidLocale = (value) => Object.prototype.hasOwnProperty.call(dictionaries, value)

const detectSystemLocale = () => {
  let raw = ''
  try {
    raw = (navigator.languages && navigator.languages[0]) || navigator.language || ''
  } catch {
    raw = ''
  }
  return String(raw).toLowerCase().startsWith('es') ? 'es' : 'en'
}

const readStoredLocale = () => {
  try {
    return localStorage.getItem(STORAGE_KEY) || ''
  } catch {
    return ''
  }
}

const state = reactive({
  locale: isValidLocale(readStoredLocale()) ? readStoredLocale() : detectSystemLocale()
})

const applyDocumentLang = (locale) => {
  try {
    document.documentElement.lang = locale
  } catch {
    /* sin document (tests) */
  }
}
applyDocumentLang(state.locale)

// Busca la clave 'a.b.c' dentro de un diccionario. Devuelve undefined si no
// está o si lo que hay en medio no es una rama.
const lookup = (dict, key) => {
  let node = dict
  for (const part of key.split('.')) {
    if (node === null || typeof node !== 'object') return undefined
    node = node[part]
  }
  return typeof node === 'string' ? node : undefined
}

// t('downloads.added') o t('sidebar.downloadsCount', { n: 3 }). Las llaves
// {nombre} se sustituyen por el parámetro con ese nombre; si falta el
// parámetro la llave se queda como estaba, que delata el error en pantalla en
// vez de romper la vista.
export function t(key, params) {
  let text = lookup(dictionaries[state.locale], key)
  if (text === undefined) text = lookup(es, key)
  if (text === undefined) return key
  if (!params) return text
  return text.replace(/\{(\w+)\}/g, (match, name) =>
    params[name] !== undefined && params[name] !== null ? String(params[name]) : match
  )
}

// Parte una traducción por el marcador {nombre} para poder meter HTML (un
// enlace, un <code>...) entre las dos mitades: splitOn('listener.helper',
// 'link') devuelve [textoAntes, textoDespués]. Sustituye el marcador por un
// carácter de control que no aparece en ningún texto y parte por él.
const SPLIT_MARK = '0000'
export function splitOn(key, mark, params = {}) {
  const parts = t(key, { ...params, [mark]: SPLIT_MARK }).split(SPLIT_MARK)
  return parts.length === 2 ? parts : [t(key, params), '']
}

// has indica si la clave existe en el idioma activo o en el fallback. Sirve
// para valores que llegan del backend (estados, categorías): si el servidor
// manda uno nuevo que todavía no tiene traducción, se muestra tal cual en
// lugar de la clave cruda.
export function has(key) {
  return (
    lookup(dictionaries[state.locale], key) !== undefined ||
    lookup(es, key) !== undefined
  )
}

export function setLocale(locale) {
  if (!isValidLocale(locale) || state.locale === locale) return
  state.locale = locale
  try {
    localStorage.setItem(STORAGE_KEY, locale)
  } catch {
    /* localStorage no disponible: el idioma vive solo en memoria */
  }
  applyDocumentLang(locale)
}

export function useI18n() {
  return {
    locale: computed(() => state.locale),
    t,
    splitOn,
    has,
    setLocale,
    availableLocales
  }
}
