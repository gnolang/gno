// card.swift draws one community share card, 1200x630 in grayscale: a
// headline and a subtitle on the left, gno.land's mark small in the bottom
// left corner, and a footer line in the bottom right. The wording comes from
// the HEAD, SUB and FOOT environment variables; see generate.sh.
//
// Usage: swift card.swift <og-gnoland.png> <Intervar.woff2> <out.png>
import AppKit
import CoreText

let args = CommandLine.arguments
let markSrc = args[1], fontPath = args[2], out = args[3]
let env = ProcessInfo.processInfo.environment
CTFontManagerRegisterFontsForURL(URL(fileURLWithPath: fontPath) as CFURL, .process, nil)

let width = 1200, height = 630
let margin: CGFloat = 96
let ctx = CGContext(data: nil, width: width, height: height, bitsPerComponent: 8, bytesPerRow: 0,
                    space: CGColorSpaceCreateDeviceGray(), bitmapInfo: CGImageAlphaInfo.none.rawValue)!
ctx.setFillColor(gray: 0, alpha: 1)
ctx.fill(CGRect(x: 0, y: 0, width: width, height: height))
ctx.interpolationQuality = .high

// inter returns Inter at size and weight, through its variable 'wght' axis.
func inter(_ size: CGFloat, weight: Double) -> CTFont {
    let base = CTFontCreateWithName("Inter" as CFString, size, nil)
    let desc = CTFontDescriptorCreateWithAttributes([kCTFontVariationAttribute: [0x77676874: weight]] as CFDictionary)
    return CTFontCreateCopyWithAttributes(base, size, nil, desc)
}

// draw writes s with its left edge at x, or its right edge at x when right.
func draw(_ s: String, _ font: CTFont, gray: CGFloat, x: CGFloat, baseline: CGFloat, right: Bool = false) {
    let attrs: [NSAttributedString.Key: Any] = [.font: font, .foregroundColor: NSColor(white: gray, alpha: 1)]
    let line = CTLineCreateWithAttributedString(NSAttributedString(string: s, attributes: attrs))
    let b = CTLineGetBoundsWithOptions(line, .useOpticalBounds)
    ctx.textPosition = CGPoint(x: (right ? x - b.width : x) - b.minX, y: baseline)
    CTLineDraw(line, ctx)
}

// Headline and subtitle, slightly above the middle.
let headSize: CGFloat = 76, subSize: CGFloat = 32
draw(env["HEAD"]!, inter(headSize, weight: 600), gray: 1, x: margin, baseline: 352)
draw(env["SUB"]!, inter(subSize, weight: 400), gray: 0.6, x: margin, baseline: 352 - headSize * 0.55 - subSize)

// gno.land's mark and wordmark, cut from og-gnoland.png (its bounding box,
// top-left origin) at a quarter of their size.
let mark = NSImage(contentsOfFile: markSrc)!.cgImage(forProposedRect: nil, context: nil, hints: nil)!
    .cropping(to: CGRect(x: 300, y: 244, width: 600, height: 143))!
let scale: CGFloat = 0.25
let markHeight = CGFloat(mark.height) * scale
ctx.draw(mark, in: CGRect(x: margin, y: 64, width: CGFloat(mark.width) * scale, height: markHeight))

// Footer, right-aligned, centred on the mark's middle; 32px #999 stays
// legible at preview size.
let foot = inter(32, weight: 400)
draw(env["FOOT"]!, foot, gray: 0.6, x: CGFloat(width) - margin,
     baseline: 64 + markHeight / 2 - CTFontGetXHeight(foot) / 2, right: true)

let png = NSBitmapImageRep(cgImage: ctx.makeImage()!).representation(using: .png, properties: [:])!
try! png.write(to: URL(fileURLWithPath: out))
