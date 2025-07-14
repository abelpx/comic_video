import React from 'react';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { Layout, Menu } from 'antd';
import GenerateComic from './pages/GenerateComic';
import GenerateTweet from './pages/GenerateTweet';
import GenerateNovel from './pages/GenerateNovel';
import VideoToAnime from './pages/VideoToAnime';

const { Header, Content } = Layout;

export default function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Header>
          <Menu theme="dark" mode="horizontal" defaultSelectedKeys={['comic']}>
            <Menu.Item key="comic"><Link to="/">漫画生成</Link></Menu.Item>
            <Menu.Item key="tweet"><Link to="/tweet">推文生成</Link></Menu.Item>
            <Menu.Item key="novel"><Link to="/novel">小说生成</Link></Menu.Item>
            <Menu.Item key="video"><Link to="/video">视频转动漫</Link></Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: 24, minHeight: '90vh' }}>
          <Routes>
            <Route path="/" element={<GenerateComic />} />
            <Route path="/tweet" element={<GenerateTweet />} />
            <Route path="/novel" element={<GenerateNovel />} />
            <Route path="/video" element={<VideoToAnime />} />
          </Routes>
        </Content>
      </Layout>
    </BrowserRouter>
  );
} 