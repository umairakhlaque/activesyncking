import { createBrowserRouter, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Devices from './pages/Devices';
import Users from './pages/Users';
import AuditLog from './pages/AuditLog';
import Policies from './pages/Policies';
import Login from './pages/Login';

function requireAuth(element: React.ReactNode): React.ReactNode {
  if (!localStorage.getItem('sg_admin_token')) {
    return <Navigate to="/login" replace />;
  }
  return element;
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: <Navigate to="/dashboard" replace />,
  },
  {
    path: '/',
    element: requireAuth(<Layout />) as React.ReactElement,
    children: [
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'devices', element: <Devices /> },
      { path: 'users', element: <Users /> },
      { path: 'audit', element: <AuditLog /> },
      { path: 'policies', element: <Policies /> },
    ],
  },
]);
