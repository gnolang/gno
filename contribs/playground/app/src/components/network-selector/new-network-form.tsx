import React, { useState } from 'react'
import { useForm } from 'react-hook-form'

import type { Chain } from '@gnostudio/core'

import { css, cx } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { button, input } from '@/styled-system/recipes'

type ChainForm = Pick<Chain, 'displayName' | 'rpcUrl'>

interface Props {
  /**
   * Form submit handler
   * @param c Created chain
   */
  onAdd?: (c: Chain) => void

  /**
   * Form validation handler. Used to identify duplicate entries.
   * @param key Chain property key (rpcUrl or displayName)
   * @param val Property value
   */
  onValidate?: (key: keyof ChainForm, val: string) => boolean

  /**
   * Chain ID resolver. Triggered before form submit to resolve chain ID.
   * Should return `null` on error.
   */
  chainIdProvider?: (rpcUrl: string) => Promise<string | null>

  /**
   * Controls whether form is disabled.
   */
  disabled?: boolean
}

const URL_REGEX = /^(http|https):\/\/[^ "]+$/

const noopValidator = () => true
const noopChainIdProvider = () => Promise.resolve('')

export const NewNetworkForm: React.FC<Props> = ({
  onAdd,
  disabled = false,
  onValidate = noopValidator,
  chainIdProvider = noopChainIdProvider,
}) => {
  const [isOpen, setOpen] = useState(false)
  const [isBusy, setBusy] = useState(false)
  const inputStyles = input()

  const {
    reset,
    register,
    handleSubmit,
    formState: { isDirty, isValid },
  } = useForm<ChainForm>()

  const onDismiss = () => {
    reset()
    setOpen(false)
  }

  const onSubmit = handleSubmit(async (data) => {
    if (isBusy) {
      return
    }

    setBusy(true)
    const id = await chainIdProvider(data.rpcUrl)
    setBusy(false)
    if (!id) {
      return
    }

    onAdd?.({
      id,
      ...data,
    })
    reset()
    setOpen(false)
  })

  return (
    <>
      <button
        className={button({ block: true })}
        data-testid="new-network-btn"
        hidden={isOpen}
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        Custom Network
      </button>
      <form onSubmit={onSubmit} hidden={!isOpen} autoComplete="off" data-testid="new-network-form">
        <div className={stack({ gap: '2' })}>
          <input
            {...register('rpcUrl', {
              disabled,
              required: true,
              maxLength: 96,
              pattern: URL_REGEX,
              validate: (val) => onValidate('rpcUrl', val),
            })}
            type="url"
            className={cx(inputStyles.root, inputStyles.input)}
            placeholder="RPC address"
            disabled={disabled || isBusy}
            spellCheck={false}
            autoComplete="off"
          />
          <input
            {...register('displayName', {
              disabled,
              required: true,
              maxLength: 64,
              validate: (val) => onValidate('displayName', val),
            })}
            type="text"
            className={cx(inputStyles.root, inputStyles.input)}
            placeholder="Display name"
            disabled={disabled || isBusy}
          />
        </div>
        <div className={cx(stack({ gap: '3' }), css({ mt: '3' }))}>
          <button type="submit" className={button({ block: true })} disabled={disabled || !isDirty || !isValid}>
            {isBusy ? 'Please wait...' : 'Add'}
          </button>
          <button
            type="button"
            data-testid="network-form-close-btn"
            className={button({ variant: 'ghost', block: true })}
            disabled={disabled || isBusy}
            onClick={onDismiss}
          >
            Cancel
          </button>
        </div>
      </form>
    </>
  )
}
