import React, { useState } from "react";

export default function Feedback() {
  const [isOpen, setIsOpen] = useState(false);
  const [isFormCallback, setIsFormCallback] = useState(false);
  const toggle = (e) => {
    e.preventDefault();
    setIsOpen(!isOpen);
  };

  const [formData, setFormData] = useState({
    email: "",
    feedback: "",
  });

  const feedbackMsg = {
    info: "Leave feedback",
    callback: "Thank you for your feedback! ❤️",
  };

  const handleInputChange = (event) => {
    event.preventDefault();
    const { name, value } = event.target;
    setFormData((prevProps) => ({
      ...prevProps,
      [name]: value,
    }));
  };

  const sendData = async (e) => {
    e.preventDefault();

    if (formData.email === "" || formData.feedback === "") return;

    const formURL = `https://docs.google.com/forms/d/188U6r1PL0zvwo5nrBwpIh8CVtgvOMdvst2YM3PWT7hw/formResponse`;
    const inputsName = {
      email: "entry.331557816",
      feedback: "entry.1077515444",
      pageId: "entry.420531539",
      pageCommit: "entry.159427272",
    };

    const formParams = new URLSearchParams();
    for (const field in inputsName) {
      if (formData[field]) formParams.append(inputsName[field], formData[field]);
    }

    formParams.append(inputsName.pageId, window.location.pathname);

    const editLink = document.querySelector(".theme-edit-this-page").getAttribute("href");
    const filePath = editLink.split("master/")[1];

    await fetch(`https://api.github.com/repos/gnolang/gno/commits?path=/${filePath}`)
      .then((res) => res.json())
      .then((data) => {
        formParams.append(inputsName.pageCommit, data[0].sha);
      })
      .catch((error) => {
        formParams.append(inputsName.pageCommit, `error: ${error}`);
        console.error("Last page commit error :", error);
      });

    const url = formURL + "?" + formParams + "&submit=Submit";

    fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      mode: "no-cors", // this line is required for CORS policy reasons -> no usefull response available
    })
      .then((_) => {
        setFormData({ email: "", feedback: "" });
        setIsOpen(false);

        setIsFormCallback(true);
        setTimeout(() => setIsFormCallback(false), 2000);
      })
      .catch((error) => {
        console.error("Error submitting form data:", error);
      });
  };

  return (
    <div className="feedback">
      <div className="footer__title">Was this page helpful?</div>
      {isOpen ? (
        <form className="feedback__form" onSubmit={sendData}>
          <button className="feedback__inner-btn feedback__close" aria-label="Close" onClick={toggle}>
            ✕
          </button>
          <div>
            <div className="feedback__field">
              <label htmlFor="feedback-email">Email</label>
              <input type="email" name="email" id="feedback-email" placeholder="example@domain.com" value={formData.email} onChange={handleInputChange} required />
            </div>
            <div className="feedback__field">
              <label htmlFor="feedback-review">Review</label>
              <textarea type="text" name="feedback" id="feedback-review" placeholder="Your review here" rows="6" value={formData.feedback} onChange={handleInputChange} required />
            </div>
            <button className="feedback__inner-btn feedback__send" type="submit">
              Send Feedback
            </button>
          </div>
        </form>
      ) : (
        <button className="feedback__btn" onClick={toggle}>
          {!isFormCallback ? feedbackMsg.info : feedbackMsg.callback}
        </button>
      )}
    </div>
  );
}
