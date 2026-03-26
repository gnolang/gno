import { defineProject } from 'vitest/config'

export default defineProject({
  test: {
    testTimeout: process.env.TEST_TIMEOUT ? parseInt(process.env.TEST_TIMEOUT, 10) : 5000,
  },
})
