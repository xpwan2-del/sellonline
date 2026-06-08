<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import { useLocalized } from '../composables/useProduct'
import { processHtmlForDisplay } from '../utils/content'
import { useAnnouncement, type HomeAnnouncement } from '../composables/useAnnouncement'

const props = defineProps<{
  announcement: HomeAnnouncement
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { dismissForever, dismissToday, closeForSession } = useAnnouncement()

const title = computed(() => getLocalizedText(props.announcement.title))

const sanitizedContent = computed(() => {
  const raw = getLocalizedText(props.announcement.content)
  const withImages = processHtmlForDisplay(String(raw || ''))
  return DOMPurify.sanitize(withImages, {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 's', 'code', 'pre', 'blockquote', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'span', 'div', 'img', 'hr', 'table', 'thead', 'tbody', 'tr', 'td', 'th', 'colgroup', 'col'],
    ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'title', 'style', 'colspan', 'rowspan', 'width'],
    ALLOW_DATA_ATTR: false,
    ALLOWED_URI_REGEXP: /^(?:https?:|mailto:|tel:|#|\/(?!\/))/i,
  })
})

const typeStyle = computed(() => {
  switch (props.announcement.type) {
    case 'warning':
      return {
        iconWrap: 'bg-warning-soft text-warning ring-warning/20',
        icon: 'M12 9v2m0 4h.01M5.07 19h13.86a2 2 0 001.74-3L13.74 4a2 2 0 00-3.48 0L3.33 16a2 2 0 001.74 3z',
      }
    case 'info':
      return {
        iconWrap: 'bg-info-soft text-info ring-info/20',
        icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
      }
    default:
      return {
        iconWrap: 'bg-primary-soft text-primary ring-primary/20',
        icon: 'M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z',
      }
  }
})

const close = () => emit('update:visible', false)

const handleClose = () => {
  closeForSession(props.announcement.version)
  close()
}

const handleDismissToday = () => {
  dismissToday(props.announcement.version)
  close()
}

const handleDismissForever = () => {
  dismissForever(props.announcement.version)
  close()
}

const dialogRef = ref<HTMLElement | null>(null)

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    handleClose()
  }
}

watch(
  () => props.visible,
  async (isVisible) => {
    if (isVisible) {
      document.addEventListener('keydown', onKeydown)
      await nextTick()
      const firstButton = dialogRef.value?.querySelector('button') as HTMLElement | null
      firstButton?.focus()
    } else {
      document.removeEventListener('keydown', onKeydown)
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-300 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="visible"
        class="fixed inset-0 z-[120] flex items-center justify-center p-4 sm:p-6"
        @click.self="handleClose"
      >
        <div class="fixed inset-0 bg-black/55 backdrop-blur-md" aria-hidden="true" />
        <Transition
          enter-active-class="transition duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)]"
          enter-from-class="opacity-0 scale-95 translate-y-3"
          enter-to-class="opacity-100 scale-100 translate-y-0"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="opacity-100 scale-100 translate-y-0"
          leave-to-class="opacity-0 scale-95 translate-y-2"
        >
          <div
            v-if="visible"
            ref="dialogRef"
            role="dialog"
            aria-modal="true"
            aria-labelledby="announcement-modal-title"
            class="theme-panel relative z-10 flex max-h-[86vh] w-full max-w-lg flex-col overflow-hidden rounded-3xl border"
            style="box-shadow: var(--ui-shadow-card)"
          >
            <!-- Header -->
            <div class="flex items-center gap-4 px-6 pb-5 pt-6">
              <div
                class="flex size-12 shrink-0 items-center justify-center rounded-2xl ring-1 ring-inset"
                :class="typeStyle.iconWrap"
              >
                <svg class="size-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" :d="typeStyle.icon" />
                </svg>
              </div>
              <h3
                id="announcement-modal-title"
                class="line-clamp-2 min-w-0 flex-1 text-lg font-semibold leading-snug theme-text-primary"
              >
                {{ title }}
              </h3>
              <button
                type="button"
                class="flex size-8 shrink-0 items-center justify-center rounded-full theme-text-muted transition-colors hover:bg-surface-soft hover:theme-text-primary"
                :aria-label="t('announcement.close')"
                @click="handleClose"
              >
                <svg class="size-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Body -->
            <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-2">
              <div class="theme-prose prose prose-sm max-w-none dark:prose-invert" v-html="sanitizedContent"></div>
            </div>

            <!-- Footer -->
            <div class="flex flex-col gap-3 border-t theme-border px-6 pb-5 pt-4 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center justify-center gap-2 text-xs sm:justify-start">
                <button
                  type="button"
                  class="rounded-md px-1 py-0.5 theme-text-muted transition-colors hover:theme-text-primary"
                  @click="handleDismissToday"
                >
                  {{ t('announcement.dismissToday') }}
                </button>
                <span class="select-none opacity-30 theme-text-muted">·</span>
                <button
                  type="button"
                  class="rounded-md px-1 py-0.5 theme-text-muted transition-colors hover:theme-text-primary"
                  @click="handleDismissForever"
                >
                  {{ t('announcement.dismissForever') }}
                </button>
              </div>
              <button
                type="button"
                class="theme-btn-primary inline-flex w-full items-center justify-center rounded-xl border px-6 py-2.5 text-sm sm:w-auto"
                @click="handleClose"
              >
                {{ t('announcement.close') }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
