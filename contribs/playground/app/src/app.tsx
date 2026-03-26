import '@fontsource-variable/inter'
import './globals.css'

import React from 'react'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { StoreProvider } from './contexts'
import { AppRouter } from './router'
import { rootStore } from './store'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
})

const store = rootStore.create()

export const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <StoreProvider value={store}>
        <AppRouter />
      </StoreProvider>
    </QueryClientProvider>
  )
}
