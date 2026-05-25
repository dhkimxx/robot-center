import { BrowserRouter } from "react-router-dom";
import { ControlCenterApp } from "./app/ControlCenterShell.jsx";

export default function App() {
  return (
    <BrowserRouter>
      <ControlCenterApp />
    </BrowserRouter>
  );
}
