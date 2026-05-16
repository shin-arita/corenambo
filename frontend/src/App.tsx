import { BrowserRouter, Route, Routes } from "react-router-dom";
import UserRegistrationPage from "./pages/UserRegistrationPage";
import UserRegistrationCompletePage from "./pages/UserRegistrationCompletePage";
import UserRegistrationVerifyPage from "./pages/UserRegistrationVerifyPage";
import UserRegistrationSuccessPage from "./pages/UserRegistrationSuccessPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/registration" element={<UserRegistrationPage />} />
        <Route path="/registration/complete" element={<UserRegistrationCompletePage />} />
        <Route path="/registration/verify" element={<UserRegistrationVerifyPage />} />
        <Route path="/registration/success" element={<UserRegistrationSuccessPage />} />
      </Routes>
    </BrowserRouter>
  );
}
