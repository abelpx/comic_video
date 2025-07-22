import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import { Toaster } from 'react-hot-toast';
import { HelmetProvider } from 'react-helmet-async';
import { motion, AnimatePresence } from 'framer-motion';

// 导入样式
import './styles/global.css';

// 导入主题
import { getTheme } from './styles/theme';
import { useUI } from './store';

// 导入布局组件
import AppLayout from './components/Layout/AppLayout';
import AuthLayout from './components/Layout/AuthLayout';

// 导入页面组件
import LandingPage from './pages/LandingPage';
import LoginPage from './pages/auth/LoginPage';
import RegisterPage from './pages/auth/RegisterPage';
import Dashboard from './pages/Dashboard';
import NovelToVideo from './pages/NovelToVideo';
import GenerateComic from './pages/GenerateComic';
import GenerateTweet from './pages/GenerateTweet';
import GenerateNovel from './pages/GenerateNovel';
import VideoToAnime from './pages/VideoToAnime';
import WorksPage from './pages/WorksPage';
import QuotaPage from './pages/QuotaPage';
import PricingPage from './pages/PricingPage';
import ProfilePage from './pages/ProfilePage';
import WorkflowPage from './pages/WorkflowPage';

// 导入路由保护组件
import ProtectedRoute from './components/ProtectedRoute';

export default function App() {
  const { theme } = useUI();
  const antdTheme = getTheme(theme === 'dark');

  return (
    <HelmetProvider>
      <ConfigProvider theme={antdTheme}>
        <AntApp>
          <BrowserRouter>
            <AnimatePresence mode="wait">
              <Routes>
                {/* 公开路由 */}
                <Route path="/" element={<LandingPage />} />
                <Route path="/pricing" element={<PricingPage />} />

                {/* 认证路由 */}
                <Route path="/auth/*" element={
                  <AuthLayout>
                    <Routes>
                      <Route path="login" element={<LoginPage />} />
                      <Route path="register" element={<RegisterPage />} />
                    </Routes>
                  </AuthLayout>
                } />

                {/* 受保护的应用路由 */}
                <Route path="/app/*" element={
                  <ProtectedRoute>
                    <AppLayout>
                      <Routes>
                        <Route path="dashboard" element={<Dashboard />} />
                        <Route path="novel-to-video" element={<NovelToVideo />} />
                        <Route path="generate-comic" element={<GenerateComic />} />
                        <Route path="generate-tweet" element={<GenerateTweet />} />
                        <Route path="generate-novel" element={<GenerateNovel />} />
                        <Route path="video-to-anime" element={<VideoToAnime />} />
                        <Route path="workflow" element={<WorkflowPage />} />
                        <Route path="works" element={<WorksPage />} />
                        <Route path="quota" element={<QuotaPage />} />
                        <Route path="profile" element={<ProfilePage />} />
                      </Routes>
                    </AppLayout>
                  </ProtectedRoute>
                } />
              </Routes>
            </AnimatePresence>
          </BrowserRouter>

          {/* 全局通知 */}
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: theme === 'dark' ? '#1e293b' : '#ffffff',
                color: theme === 'dark' ? '#f1f5f9' : '#0f172a',
                border: `1px solid ${theme === 'dark' ? '#334155' : '#e2e8f0'}`,
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
              },
            }}
          />
        </AntApp>
      </ConfigProvider>
    </HelmetProvider>
  );
}