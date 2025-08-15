<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '#ui/types'

definePageMeta({
  layout: 'auth'
})

useSeoMeta({
  title: 'Login',
  description: 'Login to your account to continue'
})

const toast = useToast()

const fields = [
  {
    name: 'email',
    type: 'text' as const,
    label: 'Email',
    placeholder: 'Enter your email',
    required: true,
    description: 'Send magic link to sign in'
  }
]

const providers = [
  {
    label: 'Google',
    icon: 'i-simple-icons-google',
    onClick: () => {
      toast.add({ title: 'Google', description: 'Login with Google' })
    }
  },
  {
    label: 'GitHub',
    icon: 'i-simple-icons-github',
    onClick: () => {
      toast.add({ title: 'GitHub', description: 'Login with GitHub' })
    }
  }
]

const schema = z.object({
  email: z.string().email('Invalid email')
})

type Schema = z.output<typeof schema>

const loading = ref(false)
const sent = ref(false)

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  loading.value = true
  try {
    // Here you would integrate with your auth service to send the magic link
    // For example: await auth.sendMagicLink(payload.data.email)
    console.log('Magic link requested for:', payload.data.email)
    sent.value = true
    toast.add({
      title: 'Check your email',
      description: 'We sent you a magic link to sign in',
      icon: 'i-lucide-mail'
    })
  } catch (error) {
    toast.add({
      title: 'Error',
      description: 'Failed to send magic link. Please try again.',
      color: 'error'
    })
    console.log(error)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <UAuthForm
    :fields="fields"
    :schema="schema"
    :providers="providers"
    :loading="loading"
    :disabled="sent"
    title="Welcome back"
    icon="i-lucide-lock"
    :submit="{ label: sent ? 'Check your email' : 'Send magic link' }"
    @submit="onSubmit"
  >
    <template #description>
      Don't have an account?
      <ULink
        to="/signup"
        class="text-primary font-medium"
      >Sign up</ULink>.
    </template>

    <template #footer>
      By signing in, you agree to our
      <ULink
        to="/"
        class="text-primary font-medium"
      >Terms of Service</ULink>.
    </template>
  </UAuthForm>
</template>
