<script setup>
import { computed, watch } from 'vue'
import { ArrowLeft, ArrowRight } from '../icons'
import { useI18n } from '../i18n'

import AppButton from './AppButton.vue'

const { t } = useI18n()

const props = defineProps({
  // Total de elementos de la lista a paginar.
  total: {
    type: Number,
    required: true
  },
  // Tamaños de página ofrecidos en el selector.
  sizes: {
    type: Array,
    required: true
  }
})

// La vista es la dueña del estado (página y tamaño persistidos o reiniciados
// al filtrar): el componente solo los edita vía v-model.
const page = defineModel('page', { type: Number, default: 1 })
const size = defineModel('size', { type: Number, default: undefined })

const maxVisiblePages = 5

const actualSize = computed(() => size.value ?? props.sizes[0])

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / actualSize.value))
)

// Al borrar elementos la última página puede desaparecer: si estábamos en
// ella, retrocedemos para no quedarnos mirando una lista vacía.
watch(totalPages, (total) => {
  if (page.value > total) page.value = total
})

// Al cambiar el tamaño, la página en la que estabas ya no significa lo mismo:
// se vuelve al principio.
watch(size, () => {
  page.value = 1
})

// Como mucho 5 botones de página alrededor de la actual, para que el paginador
// no se desborde cuando la lista es larga.
const visiblePages = computed(() => {
  const total = totalPages.value
  const max = maxVisiblePages
  if (total <= max) return Array.from({ length: total }, (_, i) => i + 1)
  const end = Math.min(total, Math.max(1, page.value - 2) + max - 1)
  const start = Math.max(1, end - max + 1)
  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

const goToPage = (target) => {
  page.value = Math.min(totalPages.value, Math.max(1, target))
}

// La clase `active` convive con `btn--active` porque el tema claro y
// darling-red personaliza el botón activo vía `.pager-btn.active`.
const pageClass = (target) => ['pager-btn', { active: target === page.value }]
</script>

<!--
  Los estilos del paginador (`.history-pager`, `.pager-btn`, `.pager-side`,
  `.pager-info`, `.pager-select`) viven en `dashboard/src/style.css` y sus
  overrides de tema en `themes-dl.css`; aquí no se duplica ninguno.
-->
<template>
  <div class="history-pager">
    <template v-if="totalPages > 1">
      <AppButton
        variant="pager"
        class="pager-btn"
        :title="t('downloads.prevPage')"
        :aria-label="t('downloads.prevPage')"
        :disabled="page === 1"
        @click="goToPage(page - 1)"
      >
        <ArrowLeft :size="14" />
      </AppButton>
      <AppButton
        v-for="target in visiblePages"
        :key="target"
        variant="pager"
        :class="pageClass(target)"
        :active="target === page"
        @click="goToPage(target)"
      >
        {{ target }}
      </AppButton>
      <AppButton
        variant="pager"
        class="pager-btn"
        :title="t('downloads.nextPage')"
        :aria-label="t('downloads.nextPage')"
        :disabled="page === totalPages"
        @click="goToPage(page + 1)"
      >
        <ArrowRight :size="14" />
      </AppButton>
    </template>

    <div class="pager-side">
      <span v-if="totalPages > 1" class="pager-info">
        {{ t('downloads.pageOf', { page, total: totalPages }) }}
      </span>
      <select
        v-model="size"
        class="pager-select"
        :title="t('downloads.perPageLabel')"
        :aria-label="t('downloads.perPageLabel')"
      >
        <option v-for="option in sizes" :key="option" :value="option">
          {{ t('downloads.perPage', { n: option }) }}
        </option>
      </select>
    </div>
  </div>
</template>
