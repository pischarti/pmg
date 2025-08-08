import type { RouterConfig } from '@nuxt/schema'

export default <RouterConfig>{
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }

    if (to.hash) {
      const getElement = (hash: string): HTMLElement | null => {
        const id = hash.startsWith('#') ? hash.slice(1) : hash
        return document.getElementById(id)
      }

      return new Promise((resolve) => {
        if (!import.meta.client) {
          resolve({ left: 0, top: 0 })
          return
        }

        let attempts = 0
        const maxAttempts = 50
        const timer = setInterval(() => {
          const el = getElement(to.hash)
          attempts += 1
          if (el) {
            clearInterval(timer)
            el.scrollIntoView({ behavior: 'smooth', block: 'start' })
            resolve(false)
          } else if (attempts >= maxAttempts) {
            clearInterval(timer)
            resolve({ left: 0, top: 0 })
          }
        }, 50)
      })
    }

    return { left: 0, top: 0 }
  }
}
