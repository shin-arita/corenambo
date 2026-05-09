import { BrowserRouter, Route, Routes } from "react-router-dom";
import UserRegistrationPage from "./pages/UserRegistrationPage";
import UserRegistrationCompletePage from "./pages/UserRegistrationCompletePage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/registration" element={<UserRegistrationPage />} />
        <Route path="/registration/complete" element={<UserRegistrationCompletePage />} />
      </Routes>
    </BrowserRouter>
  );
}
