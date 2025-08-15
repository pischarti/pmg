<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '#ui/types'

definePageMeta({
  layout: 'auth'
})

useSeoMeta({
  title: 'Sign up',
  description: 'Create an account to get started'
})

const toast = useToast()

const fields = [
  {
    name: 'email',
    type: 'text' as const,
    label: 'Email',
    placeholder: 'Enter your email',
    required: true,
    description: 'We will send you a magic link to create your account'
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
      description: 'We sent you a magic link to create your account',
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
    title="Create an account"
    :submit="{ label: sent ? 'Check your email' : 'Send magic link' }"
    @submit="onSubmit"
  >
    <template #description>
      Already have an account?
      <ULink
        to="/login"
        class="text-primary font-medium"
      >Login</ULink>.
    </template>

    <template #footer>
      By signing up, you agree to our
      <ULink
        to="/"
        class="text-primary font-medium"
      >Terms of Service</ULink>.
    </template>
  </UAuthForm>
</template>
