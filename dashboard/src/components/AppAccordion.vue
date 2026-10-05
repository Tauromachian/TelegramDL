<script setup>
import { ref } from 'vue'
import { ChevronRight } from '../icons'

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' }
})

const seccionAbierta = ref(false)
</script>

<template>
  <section class="ajuste-bloque" :class="{ abierto: seccionAbierta }">
    <button
      type="button"
      class="ajuste-cabecera"
      :aria-expanded="seccionAbierta === 'temas'"
      @click="seccionAbierta = !seccionAbierta"
    >
      <span class="ajuste-icono">
        <slot name="icon"></slot>
      </span>

      <span class="ajuste-titulo">
        <p>
          <strong>{{ title }}</strong>
        </p>
        <p>
          <small>
            {{ description }}
          </small>
        </p>
      </span>

      <ChevronRight class="ajuste-flecha" :size="16" />
    </button>

    <div v-show="seccionAbierta" class="ajuste-cuerpo">
      <slot></slot>
    </div>
  </section>
</template>

<style scoped>
.ajuste-titulo > * {
  margin: 0;
}

.ajuste-titulo strong {
  font:
    600 13px 'Space Grotesk',
    sans-serif;
}
.ajuste-titulo small {
  font-size: 11px;
  color: var(--user-text-dim);
}

.ajuste-bloque {
  border: 1px solid var(--user-border);
  border-radius: 14px;
  background: var(--user-bg-base);
  overflow: hidden;
  transition: border-color 0.2s;
}

.ajuste-bloque.abierto {
  border-color: var(--user-border-light);
}

.ajuste-flecha {
  color: var(--user-text-dim);
  flex: none;
  transition:
    transform 0.2s ease,
    color 0.2s;

  margin-left: auto;
}

.ajuste-bloque.abierto .ajuste-flecha {
  transform: rotate(90deg);
  color: var(--user-accent);
}
</style>
