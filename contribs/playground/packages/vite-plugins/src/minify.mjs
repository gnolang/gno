import JavascriptObfuscator from 'javascript-obfuscator'
import { minify } from 'terser'

const MIN_COMPRESS_PASSES = 2
const MAX_COMPRESS_PASSES = 10

export function minifyPlugin(skipObfuscation = false) {
  if (skipObfuscation) {
    console.log('[minifyPlugin]: Skipping obfuscation')
  } else {
    console.log('[minifyPlugin]: Enabling obfuscation')
  }

  return {
    name: 'minify-bundle',
    async generateBundle(_, bundle) {
      for (const asset of Object.values(bundle)) {
        if (asset.type === 'chunk') {
          const minified = await minify(asset.code, {
            compress: {
              passes: skipObfuscation ? MIN_COMPRESS_PASSES : MAX_COMPRESS_PASSES,
              drop_debugger: true,
            },
            mangle: {
              toplevel: true,
              module: true,
            },
            format: {
              comments: 'some',
              preamble: `
/*
 * @preserve
 *
 * Copyright (c) 2022. All rights reserved.
 * 
 * Project Owner:
 * NewTendermint, LLC
 * 
 * Project Maintainer:
 * İlker Göktuğ ÖZTÜRK. <ilker@ilgooz.com>, <ilkergoktugozturk@gmail.com>
 * 
 * This Project, including, but not limited to, its source code, software,
 * components, any related works, associated materials, ideas, design, and
 * documentation, whether in whole, in part, or in any form, is strictly private,
 * proprietary, and confidential.
 * 
 * Any possession, access, viewing, use, copying, forking, distribution,
 * reproduction, modification, reverse engineering, disassembly, or creation of
 * derivative works of this Project, including, but not limited to, its source
 * code, software, components, any related works, associated materials, ideas,
 * design, and documentation, whether in whole, in part, or in any form, is
 * strictly prohibited regardless of intent, purpose, or authorization.
 * 
 * If you have acquired or otherwise obtained possession of, access to, or copies of
 * this Project, including but not limited to its source code, software, components,
 * related works, associated materials, ideas, designs, or documentation, whether
 * intentionally, unintentionally, authorized, or unauthorized, in whole or in part,
 * you are required to immediately cease all use, relinquish possession, and
 * permanently destroy all copies thereof. Such acquisition, possession, or access
 * does not convey or imply any license, permission, authorization, or rights
 * whatsoever to use, reproduce, or exploit the Project or any portion thereof for
 * any purpose or in any manner.
 * 
 * THIS PROJECT AND THE WORKS AVAILABLE THROUGH THIS PROJECT ARE PROVIDED “AS IS”
 * AND WITHOUT WARRANTY OF ANY KIND. IN NO EVENT SHALL THE OWNER OR MAINTAINER OF
 * THIS PROJECT BE LIABLE TO YOU OR ANY THIRD PARTY FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THIS PROJECT OR THE WORKS AVAILABLE THROUGH THIS
 * PROJECT. YOU AGREED TO INDEMNIFY, DEFEND AND HOLD THE OWNER AND MAINTAINER FROM
 * AND AGAINST ANY CLAIMS, LOSSES OR DAMAGES. 
 * 
 * This Proprietary and Confidential Notice is subject to change at any time by the
 * Project Owner or Project Maintainer. The most recent version of this Proprietary
 * and Confidential Notice shall supersede and govern all matters related to this
 * Project, its associated works, and any prior versions of the Project retroactively.
 * 
 * The Project Owner and Maintainer reserve the right and have full and exclusive
 * authority to change, modify, update, or replace this Proprietary and Confidential
 * Notice at any time, in any manner of their choosing, without seeking or requiring the
 * consent, approval, or input of any contributor, user, or other party.
 *
 * In the event of any conflict, overlap, or inconsistency between this Proprietary and
 * Confidential Notice and any other version of the Proprietary and Confidential Notice applicable
 * to private, internal, or non-public portions of the Project or Project Assets, the
 * non-public version shall prevail and shall supersede and govern to the extent of such
 * conflict, overlap, or inconsistency.
 *
 * This Proprietary and Confidential Notice governs only those portions of the Project or
 * Project Assets that are publicly visible, accessible, or available, and does not affect or
 * modify any terms applicable to any private, internal, or non-public content governed by
 * a separate or non-public version of the Proprietary and Confidential Notice.
*/
`,
            },
          })

          if (!minified.code) return

          const processedCode = skipObfuscation
            ? minified.code
            : JavascriptObfuscator.obfuscate(minified.code, {
                compact: true,
                controlFlowFlattening: true,
                controlFlowFlatteningThreshold: 0,
                simplify: true,
                target: 'browser-no-eval',
                transformObjectKeys: true,
              }).getObfuscatedCode()

          asset.code = processedCode
        }
      }
    },
  }
}
