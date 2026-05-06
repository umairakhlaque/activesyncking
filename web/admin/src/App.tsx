import { createBrowserRouter, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Devices from './pages/Devices';
import Users from './pages/Users';
import AuditLog from './pages/AuditLog';
import Policies from './pages/Policies';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to="/dashboard" replace />,
  },
  {
    path: '/',
    element: <Layout />,
    children: [
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'devices', element: <Devices /> },
      { path: 'users', element: <Users /> },
      { path: 'audit', element: <AuditLog /> },
      { path: 'policies', element: <Policies /> },
    ],
  },
]);
