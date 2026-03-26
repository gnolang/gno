import { BrowserRouter, Route, Routes } from 'react-router-dom'

import { Layout } from './components/layout'
import { FilesImport, Playground, Privacy, Terms } from './pages'
import { Embed } from './pages/embed'

export const AppRouter = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout pageType="play" />}>
          <Route path="/github/:owner/:repo" element={<FilesImport strategy="github" />} />
          <Route path="/p/:id?" element={<Playground />} />
          <Route path="/" element={<Playground />} />
        </Route>
        <Route element={<Layout pageType="standalone" />}>
          <Route path="/terms" element={<Terms />} />
          <Route path="/privacy" element={<Privacy />} />
        </Route>
        <Route path="/embed/:id?" element={<Embed />} />
      </Routes>
    </BrowserRouter>
  )
}
