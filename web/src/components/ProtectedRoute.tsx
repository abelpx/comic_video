import React, { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { useAuth } from '../store';
import { authApi } from '../services/api';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated, token, login, logout } = useAuth();
  const location = useLocation();
  const [loading, setLoading] = React.useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      if (token && !isAuthenticated) {
        try {
          // 验证token有效性
          const userProfile = await authApi.getProfile();
          login(userProfile, token);
        } catch (error) {
          // token无效，清除登录状态
          logout();
        }
      }
      setLoading(false);
    };

    checkAuth();
  }, [token, isAuthenticated, login, logout]);

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        flexDirection: 'column',
        gap: 16
      }}>
        <Spin size="large" />
        <div style={{ color: '#64748b' }}>验证登录状态...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    // 保存当前路径，登录后重定向
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;
