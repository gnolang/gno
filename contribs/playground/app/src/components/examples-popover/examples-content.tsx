import React, { useEffect, useRef, useState } from 'react'
import { PiFiles, PiGear, PiMagnifyingGlass, PiX } from 'react-icons/pi'

import { observer } from 'mobx-react-lite'

import { categories as originalCategories, type ExampleCategory, type ExampleItem } from '@/generated/examples'
import { css, cx } from '@/styled-system/css'
import { stack } from '@/styled-system/patterns'
import { button, input } from '@/styled-system/recipes'

import { useExamplesUtilities } from './use-examples-utilities'

interface ExamplesContentProps {
  onExampleSelected: () => void
}

export const ExamplesContent = observer(({ onExampleSelected }: ExamplesContentProps) => {
  const [searchText, setSearchText] = useState('')
  const [debouncedSearchText, setDebouncedSearchText] = useState('')
  const searchInputRef = useRef<HTMLInputElement>(null)
  const { reorganizeCategories, filterExamples, handleLoadExample, formatItemTitle } = useExamplesUtilities()

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchText(searchText)
    }, 300)

    return () => clearTimeout(timer)
  }, [searchText])

  useEffect(() => {
    if (searchInputRef.current) {
      searchInputRef.current.focus()
    }
  }, [])

  const categories = reorganizeCategories(originalCategories)
  const filteredCategories = categories
    .map((category: ExampleCategory) => filterExamples(category, debouncedSearchText))
    .filter((category: ExampleCategory) => category.items.length > 0)

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchText(e.target.value)
  }

  const handleExampleClick = (item: ExampleItem) => {
    handleLoadExample(item)
    onExampleSelected()
  }

  const inputStyles = input({ size: 'md' })

  return (
    <div className={css({ mt: '4', display: 'flex', flexDirection: 'column', flex: '1' })}>
      {/* Search Section */}
      <div className={css({ px: '5', pb: '4', borderBottomWidth: '1px', borderColor: 'border' })}>
        <div className={css({ position: 'relative', maxW: '400px', mx: 'auto' })}>
          <div
            className={css({
              position: 'absolute',
              left: '3',
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'foreground.muted',
              pointerEvents: 'none',
              zIndex: 1,
            })}
          >
            <PiMagnifyingGlass size={18} />
          </div>

          <input
            ref={searchInputRef}
            type="text"
            placeholder="Search examples..."
            value={searchText}
            onChange={handleSearch}
            className={cx(
              inputStyles.root,
              css({
                width: 'full',
                pl: '11',
                pr: searchText ? '11' : '4',
                py: '3',
                fontSize: 'md',
                borderColor: 'border',
                borderRadius: 'lg',
                _placeholder: { color: 'foreground.muted' },
                _focus: {
                  outline: 'none',
                  borderColor: 'gray.400',
                  boxShadow: '0 0 0 3px token(colors.gray.200/50)',
                },
              }),
            )}
          />

          {searchText && (
            <button
              onClick={() => setSearchText('')}
              className={css({
                position: 'absolute',
                right: '3',
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'foreground.muted',
                p: '1',
                borderRadius: 'md',
                _hover: { color: 'foreground', bg: 'gray.100' },
              })}
            >
              <PiX size={16} />
            </button>
          )}
        </div>
      </div>

      {/* Content Section */}
      <div className={css({ py: '6', px: '5', maxH: '500px', overflowY: 'auto', flex: '1' })}>
        {filteredCategories.length === 0 ? (
          <div
            className={stack({
              alignItems: 'center',
              justify: 'center',
              minH: '300px',
              gap: '4',
              textAlign: 'center',
            })}
          >
            <div className={css({ p: '4', borderRadius: 'full', bg: 'gray.100', color: 'foreground.muted' })}>
              <PiMagnifyingGlass size={24} />
            </div>
            <div className={stack({ gap: '2' })}>
              <h3 className={css({ fontWeight: 'semibold', color: 'foreground', fontSize: 'lg' })}>
                No examples found
              </h3>
              <p className={css({ fontSize: 'sm', color: 'foreground.muted', maxW: '300px' })}>
                Try adjusting your search query
              </p>
            </div>
            <button onClick={() => setSearchText('')} className={cx(button({ variant: 'outline', size: 'sm' }))}>
              Clear search
            </button>
          </div>
        ) : (
          <div className={stack({ gap: '8', alignItems: 'stretch' })}>
            {filteredCategories.map((category: ExampleCategory) => (
              <section key={category.title} className={css({ width: 'full' })}>
                {/* Category Header */}
                <h2 className={css({ fontSize: 'lg', fontWeight: 'medium', color: 'foreground', mb: '4' })}>
                  {category.title}
                </h2>

                {/* Examples Grid */}
                <div
                  className={css({
                    display: 'grid',
                    gridTemplateColumns: 'repeat(4, 1fr)',
                    gap: '3',
                  })}
                >
                  {category.items.map((item: ExampleItem) => (
                    <button
                      key={String(item.title) + (item.file ? String(item.file) : '')}
                      onClick={() => handleExampleClick(item)}
                      className={css({
                        display: 'flex',
                        alignItems: 'center',
                        gap: '3',
                        p: '4',
                        borderRadius: 'lg',
                        textAlign: 'left',
                        width: 'full',
                        bg: 'header',
                        borderWidth: '1px',
                        borderColor: 'border',
                        transition: 'all 0.2s ease',
                        cursor: 'pointer',
                        _hover: {
                          bg: { base: 'gray.300', _dark: 'gray.400' },
                        },
                        _focus: {
                          outline: 'none',
                          borderColor: 'border',
                        },
                      })}
                    >
                      <PiGear size={20} className={css({ color: 'foreground.muted', flexShrink: 0 })} />

                      <div className={css({ flex: '1', minW: '0' })}>
                        <h3
                          className={css({
                            fontWeight: 'medium',
                            color: 'foreground',
                            fontSize: 'sm',
                            lineHeight: 'tight',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          })}
                        >
                          {formatItemTitle(item.title)}
                        </h3>
                      </div>

                      {(item as any).isMultiFile && (
                        <PiFiles size={16} className={css({ color: 'foreground.muted', flexShrink: 0 })} />
                      )}
                    </button>
                  ))}
                </div>
              </section>
            ))}
          </div>
        )}
      </div>
    </div>
  )
})
